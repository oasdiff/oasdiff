package diff

import (
	"cmp"
	"encoding/binary"
	"maps"
	"reflect"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// The schema diff is computed as a graph shaped like the schema graph: one
// node per pair of schema values, linked to the nodes of its sub-schema
// pairs, cyclic where the schemas are. Each pair is diffed once, and a node
// is a function of its pair alone: every comparison whose answer can depend
// on the path (matching inline sub-schemas, reconciling inline and $ref
// branches, recognizing a wrapping) is left to the unroll, and the pairs it
// compares are built with the node (candidates), so the graph is complete
// before any unroll. A pair reached from outside the schema graph is
// returned unrolled: a walk that stops where a node reappears on its own
// path, decides those comparisons with the path known, and drops the nodes
// with no change below them.
//
// What a walk produces below a node depends on which of the nodes being
// expanded above it (the path) the node can reach, and those are the ones
// in its strongly connected component (cutKey). A node that reaches none of
// them has a diff of its own, computed once by a walk of its own and shared
// wherever the node is reached, as a $ref'd schema is shared by everything
// that refers to it. A node that does reach one is inside the component
// being expanded: it is expanded once per walk, where it is first reached,
// and contributes nothing where the same walk reaches it again, since the
// changes below it are already reported. So a walk costs the size of what it
// reaches, however many paths lead through it, and, being ordered, reports
// each change at the same position on every run.
//
// A comparison made while unrolling (unroller.diff) is a walk of its own,
// sharing the path; its answer is a function of the pair and the cut key.

// valuePair identifies a comparison of two schema values, independent of the
// SchemaRef wrappers it arrived through.
type valuePair struct {
	value1 *openapi3.Schema
	value2 *openapi3.Schema
}

type schemaGraph struct {
	// nodes maps each pair of schema values to its node.
	nodes map[valuePair]*SchemaDiff

	// depth is the number of schema diffs on the current stack; zero when a
	// pair is reached from outside the schema graph.
	depth int

	// candidates holds, per node, the nodes of the pairs the unroll compares
	// to decide the node's sub-schema matching and wrapping.
	candidates map[*SchemaDiff][]*SchemaDiff

	// component numbers each node's strongly connected component, from 1
	// (assignComponents).
	component  map[*SchemaDiff]int
	components int

	// unrolled memoizes the diff of each node a walk started from, by cut
	// key.
	unrolled map[*SchemaDiff]map[string]*SchemaDiff
}

func newSchemaGraph() schemaGraph {
	return schemaGraph{
		nodes:      map[valuePair]*SchemaDiff{},
		candidates: map[*SchemaDiff][]*SchemaDiff{},
		component:  map[*SchemaDiff]int{},
		unrolled:   map[*SchemaDiff]map[string]*SchemaDiff{},
	}
}

func getSchemaDiff(config *Config, state *state, schema1, schema2 *openapi3.SchemaRef) (*SchemaDiff, error) {

	if schema1 == nil || schema2 == nil || schema1.Value == nil || schema2.Value == nil {
		return getSchemaDiffInternal(config, state, schema1, schema2)
	}

	node, err := getSchemaDiffNode(config, state, schema1, schema2)
	if err != nil {
		return nil, err
	}
	if state.graph.depth == 0 {
		// reached from outside the schema graph (a media type, a parameter, a
		// component entry): the graph below the node is complete
		u := newUnroller(config, state)
		unrolled := u.unroll(node)
		return unrolled, u.err
	}
	return node, nil
}

// getSchemaDiffNode returns the node of a pair of schema values. A pair
// reached again while its own diff is still in progress (a cycle) gets the
// node in progress, so the graph links back to it the way the schemas do.
// Every child a node has is linked, changed or not: whether anything changed
// below a node is decided when the graph is unrolled, because inside a cycle
// it cannot be decided any earlier.
func getSchemaDiffNode(config *Config, state *state, schema1, schema2 *openapi3.SchemaRef) (*SchemaDiff, error) {
	graph := &state.graph
	pair := valuePair{schema1.Value, schema2.Value}
	if node, ok := graph.nodes[pair]; ok {
		return node, nil
	}

	node := &SchemaDiff{Base: schema1.Value, Revision: schema2.Value}
	graph.nodes[pair] = node
	graph.depth++
	diff, err := getSchemaDiffInternal(config, state, schema1, schema2)
	if err == nil {
		*node = *diff
		graph.candidates[node], err = getCandidates(config, state, node)
	}
	graph.depth--
	if err != nil {
		delete(graph.nodes, pair)
		return nil, err
	}
	return node, nil
}

// getCandidates builds the nodes of the pairs the unroll compares to decide
// the node's sub-schema matching (every inline pair and every pair across
// the inline/$ref boundary of a pending sub-schema list) and its wrapping
// (each side against the other side's oneOf branches).
func getCandidates(config *Config, state *state, node *SchemaDiff) ([]*SchemaDiff, error) {
	var pairs [][2]*openapi3.SchemaRef
	fields := reflect.ValueOf(node).Elem()
	for _, index := range childFields {
		subschemas, ok := fields.Field(index).Interface().(*SubschemasDiff)
		if !ok || subschemas == nil || !subschemas.pending {
			continue
		}
		for _, schemaRef1 := range subschemas.base {
			for _, schemaRef2 := range subschemas.revision {
				if isSchemaInline(schemaRef1) && isSchemaInline(schemaRef2) {
					pairs = append(pairs, [2]*openapi3.SchemaRef{schemaRef1, schemaRef2})
				}
				if isInlineRefactorBoundary(schemaRef1, schemaRef2) {
					pairs = append(pairs, [2]*openapi3.SchemaRef{trueSchemaAsEmpty(schemaRef1), trueSchemaAsEmpty(schemaRef2)})
				}
			}
		}
	}
	for _, alt := range node.Revision.OneOf {
		pairs = append(pairs, [2]*openapi3.SchemaRef{trueSchemaAsEmpty(&openapi3.SchemaRef{Value: node.Base}), trueSchemaAsEmpty(alt)})
	}
	for _, alt := range node.Base.OneOf {
		pairs = append(pairs, [2]*openapi3.SchemaRef{trueSchemaAsEmpty(alt), trueSchemaAsEmpty(&openapi3.SchemaRef{Value: node.Revision})})
	}

	var candidates []*SchemaDiff
	for _, pair := range pairs {
		if pair[0] == nil || pair[1] == nil || pair[0].Value == nil || pair[1].Value == nil {
			continue
		}
		candidate, err := getSchemaDiffNode(config, state, pair[0], pair[1])
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

// children returns the nodes an unroll of the node can reach directly.
func (graph *schemaGraph) children(node *SchemaDiff) []*SchemaDiff {
	children := slices.Clone(graph.candidates[node])
	fields := reflect.ValueOf(node).Elem()
	for _, index := range childFields {
		switch child := fields.Field(index).Interface().(type) {
		case *SchemaDiff:
			if child != nil {
				children = append(children, child)
			}
		case *SchemasDiff:
			if child != nil {
				for _, node := range child.Modified {
					children = append(children, node)
				}
			}
		case *SubschemasDiff:
			if child != nil {
				for _, modified := range child.Modified {
					children = append(children, modified.Diff)
				}
			}
		}
	}
	return children
}

// assignComponents numbers the strongly connected components of the nodes
// that have none yet (Tarjan). A node only ever links to nodes that exist
// when it is built, so a node added later cannot join an existing component,
// and the numbering of the existing nodes stands.
func (graph *schemaGraph) assignComponents() {
	index := map[*SchemaDiff]int{}
	lowLink := map[*SchemaDiff]int{}
	onStack := map[*SchemaDiff]bool{}
	var stack []*SchemaDiff

	var visit func(node *SchemaDiff)
	visit = func(node *SchemaDiff) {
		index[node] = len(index)
		lowLink[node] = index[node]
		stack = append(stack, node)
		onStack[node] = true

		for _, child := range graph.children(node) {
			if _, done := graph.component[child]; done {
				continue
			}
			if _, ok := index[child]; !ok {
				visit(child)
				lowLink[node] = min(lowLink[node], lowLink[child])
			} else if onStack[child] {
				lowLink[node] = min(lowLink[node], index[child])
			}
		}

		if lowLink[node] == index[node] {
			graph.components++
			for {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[top] = false
				graph.component[top] = graph.components
				if top == node {
					break
				}
			}
		}
	}
	for _, node := range graph.nodes {
		if _, done := graph.component[node]; !done {
			visit(node)
		}
	}
}

// unroller walks the graph from a node and produces its diff, expanding
// each node once.
type unroller struct {
	config *Config
	state  *state

	// stack holds the nodes being expanded, and path their positions in it.
	// A comparison made during the walk shares them (diff).
	stack []*SchemaDiff
	path  map[*SchemaDiff]int

	// visited holds the nodes this walk has expanded; empty before the walk
	// starts.
	visited map[*SchemaDiff]bool

	// err is the first error a comparison made during the unroll returned.
	err error
}

func newUnroller(config *Config, state *state) *unroller {
	return &unroller{config: config, state: state, path: map[*SchemaDiff]int{}, visited: map[*SchemaDiff]bool{}}
}

func (u *unroller) unroll(node *SchemaDiff) *SchemaDiff {
	if node == nil {
		return nil
	}
	if _, ok := u.path[node]; ok {
		return nil
	}
	key := u.cutKey(node)
	if key != "" && len(u.visited) > 0 {
		// inside the component this walk is expanding
		if u.visited[node] {
			return nil
		}
		return u.expand(node)
	}
	return u.walk(node, key)
}

// walk returns the diff of the node as a walk of its own, computed once per
// cut key.
func (u *unroller) walk(node *SchemaDiff, key string) *SchemaDiff {
	graph := &u.state.graph
	if unrolled, ok := graph.unrolled[node][key]; ok {
		return unrolled
	}
	walk := &unroller{config: u.config, state: u.state, stack: u.stack, path: u.path, visited: map[*SchemaDiff]bool{}}
	unrolled := walk.expand(node)
	if walk.err != nil {
		u.fail(walk.err)
	}
	if graph.unrolled[node] == nil {
		graph.unrolled[node] = map[string]*SchemaDiff{}
	}
	graph.unrolled[node][key] = unrolled
	return unrolled
}

func (u *unroller) expand(node *SchemaDiff) *SchemaDiff {
	u.visited[node] = true
	u.path[node] = len(u.stack)
	u.stack = append(u.stack, node)
	unrolled := u.copy(node)
	u.stack = u.stack[:len(u.stack)-1]
	delete(u.path, node)
	return unrolled
}

// cutKey identifies the nodes on the path that a walk from the node cuts:
// those in the node's strongly connected component. Empty when there are
// none, as for a node reached from outside the schema graph.
func (u *unroller) cutKey(node *SchemaDiff) string {
	graph := &u.state.graph
	component, ok := graph.component[node]
	if !ok {
		graph.assignComponents()
		if component, ok = graph.component[node]; !ok {
			return ""
		}
	}
	var cut []*SchemaDiff
	for _, ancestor := range u.stack {
		if graph.component[ancestor] == component {
			cut = append(cut, ancestor)
		}
	}
	if len(cut) == 0 {
		return ""
	}
	slices.SortFunc(cut, func(a, b *SchemaDiff) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})
	key := make([]byte, 0, 8*len(cut))
	for _, ancestor := range cut {
		key = binary.AppendUvarint(key, uint64(reflect.ValueOf(ancestor).Pointer()))
	}
	return string(key)
}

func (u *unroller) copy(node *SchemaDiff) *SchemaDiff {
	unrolled := *node
	fields := reflect.ValueOf(&unrolled).Elem()
	for _, index := range childFields {
		field := fields.Field(index)
		switch child := field.Interface().(type) {
		case *SchemaDiff:
			field.Set(reflect.ValueOf(u.unroll(child)))
		case *SchemasDiff:
			field.Set(reflect.ValueOf(u.schemas(child)))
		case *SubschemasDiff:
			field.Set(reflect.ValueOf(u.subschemas(child)))
		}
	}
	unrolled.OneOfWrappingDiff = getOneOfWrappingDiff(u, node.Base, node.Revision)
	unrolled.NullableWrappingDiff = getNullableWrappingDiff(u, node.Base, node.Revision)
	if unrolled.Empty() {
		return nil
	}
	return &unrolled
}

// childFields indexes the fields of SchemaDiff that hold schema diffs, found
// by type so that a new such field is unrolled without being listed here.
// TestUnroll_CutsEveryChildField fails for a field that holds schema diffs in
// a form the unroll does not follow.
var childFields = findChildFields()

func findChildFields() []int {
	var fields []int
	for field := range reflect.TypeFor[SchemaDiff]().Fields() {
		switch field.Type {
		case reflect.TypeFor[*SchemaDiff](), reflect.TypeFor[*SchemasDiff](), reflect.TypeFor[*SubschemasDiff]():
			fields = append(fields, field.Index[0])
		}
	}
	return fields
}

func (u *unroller) schemas(diff *SchemasDiff) *SchemasDiff {
	if diff == nil {
		return nil
	}
	unrolled := *diff
	unrolled.Modified = ModifiedSchemasMap{}
	// in name order, so that which position expands a node first is the same
	// on every run
	for _, name := range slices.Sorted(maps.Keys(diff.Modified)) {
		if child := u.unroll(diff.Modified[name]); child != nil {
			unrolled.Modified[name] = child
		}
	}
	if unrolled.Empty() {
		return nil
	}
	return &unrolled
}

func (u *unroller) subschemas(diff *SubschemasDiff) *SubschemasDiff {
	if diff == nil {
		return nil
	}
	unrolled := &SubschemasDiff{
		Added:    slices.Clone(diff.Added),
		Deleted:  slices.Clone(diff.Deleted),
		Modified: ModifiedSubschemas{},
	}
	for _, modified := range diff.Modified {
		if child := u.unroll(modified.Diff); child != nil {
			unrolled.Modified = append(unrolled.Modified, &ModifiedSubschema{Base: modified.Base, Revision: modified.Revision, Diff: child})
		}
	}
	if diff.pending {
		inline, err := getSubschemasInlineDiff(u, diff.base, diff.revision)
		if err != nil {
			u.fail(err)
			return nil
		}
		unrolled = reconcileInlineRefRefactors(u, unrolled.combine(inline), diff.base, diff.revision)
	}
	if unrolled.Empty() {
		return nil
	}
	return unrolled
}

func (u *unroller) fail(err error) {
	if u.err == nil {
		u.err = err
	}
}

// diff returns the diff of a pair as a walk of its own, cut where it reaches
// a node this walk is still expanding, building the pair's node if the graph
// does not have it yet. schemaRef1 is from the base and schemaRef2 from the
// revision, like every node: a pair the other way round would be a node the
// path never holds, so nothing would cut it.
func (u *unroller) diff(schemaRef1, schemaRef2 *openapi3.SchemaRef) (*SchemaDiff, error) {
	if schemaRef1 == nil || schemaRef2 == nil || schemaRef1.Value == nil || schemaRef2.Value == nil {
		return getSchemaDiffInternal(u.config, u.state, schemaRef1, schemaRef2)
	}
	node, err := getSchemaDiffNode(u.config, u.state, schemaRef1, schemaRef2)
	if err != nil {
		return nil, err
	}
	if _, ok := u.path[node]; ok {
		return nil, nil
	}
	return u.walk(node, u.cutKey(node)), nil
}

// identical reports whether a pair diffs to nothing on the current path.
func (u *unroller) identical(schemaRef1, schemaRef2 *openapi3.SchemaRef) (bool, error) {
	diff, err := u.diff(schemaRef1, schemaRef2)
	return diff == nil, err
}

// equivalent reports whether a pair has the same validation contract on the
// current path: annotation-only differences are ignored (see
// SchemaRefsValidationEquivalent).
func (u *unroller) equivalent(schemaRef1, schemaRef2 *openapi3.SchemaRef) bool {
	diff, err := u.diff(trueSchemaAsEmpty(schemaRef1), trueSchemaAsEmpty(schemaRef2))
	if err != nil {
		u.fail(err)
		return false
	}
	return !schemaDiffHasValidationChanges(diff)
}
