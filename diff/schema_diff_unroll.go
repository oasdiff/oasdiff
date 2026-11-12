package diff

import (
	"cmp"
	"encoding/binary"
	"maps"
	"reflect"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

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
		graph.assignComponents(node)
		component = graph.component[node]
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
	unrolled := node.withChildren(u)
	// recognizing a wrapping compares schema pairs that can cycle back to
	// this node, so it is decided here, with the path known
	unrolled.OneOfWrappingDiff = getOneOfWrappingDiff(u, node.Base, node.Revision)
	unrolled.NullableWrappingDiff = getNullableWrappingDiff(u, node.Base, node.Revision)
	if unrolled.Empty() {
		return nil
	}
	return &unrolled
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
