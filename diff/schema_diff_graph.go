package diff

import (
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

	node := &SchemaDiff{}
	graph.nodes[pair] = node
	graph.depth++
	diff, err := getSchemaDiffInternal(config, state, schema1, schema2)
	if err == nil {
		*node = *diff
		// the copy carries the compared schemas; the names come from the
		// documents, which the comparison itself does not see
		node.BaseComponent = componentName(state.baseComponents, schema1)
		node.RevisionComponent = componentName(state.revisionComponents, schema2)
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
	for _, subschemas := range node.pendingSubschemas() {
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
	return append(slices.Clone(graph.candidates[node]), node.children()...)
}

// assignComponents numbers the strongly connected components of the nodes
// reachable from root that have none yet (Tarjan). A node only ever links to
// nodes that exist when it is built, so a node added later cannot join an
// existing component, and the numbering of the existing nodes stands. It
// visits only what root reaches, so its cost is the new nodes rather than the
// whole graph.
func (graph *schemaGraph) assignComponents(root *SchemaDiff) {
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
	visit(root)
}
