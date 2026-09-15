package diff

import (
	"math"

	"github.com/getkin/kin-openapi/openapi3"
)

// The schema diff is computed as a graph shaped like the schema graph: one
// node per pair of schema values, linked to the nodes of its sub-schema
// pairs, cyclic where the schemas are. Each pair is diffed once. A pair
// reached from outside the schema graph is returned unrolled: a walk that
// stops where a node reappears on its own path, drops the nodes with no
// change below them, and shares the unrolled nodes that do not depend on
// the path.

// valuePair identifies a comparison of two schema values, independent of the
// SchemaRef wrappers it arrived through.
type valuePair struct {
	value1 *openapi3.Schema
	value2 *openapi3.Schema
}

type schemaGraph struct {
	// nodes maps each pair of schema values to its node.
	nodes map[valuePair]*SchemaDiff

	// inProgress holds the nodes whose diff is being computed on the current
	// stack; their fields are not filled yet.
	inProgress map[*SchemaDiff]struct{}

	// depth is the number of schema diffs on the current stack; zero when a
	// pair is reached from outside the schema graph.
	depth int

	// unrolled memoizes the unrolled diff of each node whose unrolling did
	// not cut into a node above it, so it is the same on every path.
	unrolled map[*SchemaDiff]*SchemaDiff
}

func newSchemaGraph() schemaGraph {
	return schemaGraph{
		nodes:      map[valuePair]*SchemaDiff{},
		inProgress: map[*SchemaDiff]struct{}{},
		unrolled:   map[*SchemaDiff]*SchemaDiff{},
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
		return state.graph.unroll(node), nil
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
	graph.inProgress[node] = struct{}{}
	graph.depth++
	diff, err := getSchemaDiffInternal(config, state, schema1, schema2)
	graph.depth--
	delete(graph.inProgress, node)
	if err != nil {
		delete(graph.nodes, pair)
		return nil, err
	}

	*node = *diff
	return node, nil
}

// unroll returns the diff below a node, nil when nothing changed. A node
// reached again on its own path is a cycle and contributes nothing there:
// every change below it is already reported where it was first reached. A
// node whose diff is still in progress contributes nothing either, since its
// fields are not filled yet.
func (graph *schemaGraph) unroll(node *SchemaDiff) *SchemaDiff {
	u := unroller{graph: graph, path: map[*SchemaDiff]int{}, minCut: math.MaxInt}
	return u.unroll(node)
}

type unroller struct {
	graph *schemaGraph

	// path maps each node on the current unrolling path to its depth.
	path map[*SchemaDiff]int

	// minCut is the shallowest depth a cut has targeted since the current
	// node began; -1 for a node in progress, math.MaxInt while none has.
	minCut int
}

func (u *unroller) unroll(node *SchemaDiff) *SchemaDiff {
	if node == nil {
		return nil
	}
	if _, ok := u.graph.inProgress[node]; ok {
		u.minCut = -1
		return nil
	}
	if depth, ok := u.path[node]; ok {
		u.minCut = min(u.minCut, depth)
		return nil
	}
	if unrolled, ok := u.graph.unrolled[node]; ok {
		return unrolled
	}

	depth := len(u.path)
	u.path[node] = depth
	defer delete(u.path, node)

	outerMinCut := u.minCut
	u.minCut = math.MaxInt
	unrolled := u.copy(node)
	// A cut into this node itself falls at the same place on every path, so
	// the unrolled diff is a function of the node alone; a cut into a node
	// above it makes the unrolled diff depend on the path.
	if u.minCut >= depth {
		u.graph.unrolled[node] = unrolled
	}
	u.minCut = min(outerMinCut, u.minCut)
	return unrolled
}

func (u *unroller) copy(node *SchemaDiff) *SchemaDiff {
	unrolled := *node
	unrolled.OneOfDiff = u.subschemas(node.OneOfDiff)
	unrolled.AnyOfDiff = u.subschemas(node.AnyOfDiff)
	unrolled.AllOfDiff = u.subschemas(node.AllOfDiff)
	unrolled.NotDiff = u.unroll(node.NotDiff)
	unrolled.ItemsDiff = u.unroll(node.ItemsDiff)
	unrolled.PropertiesDiff = u.schemas(node.PropertiesDiff)
	unrolled.AdditionalPropertiesDiff = u.unroll(node.AdditionalPropertiesDiff)
	unrolled.PrefixItemsDiff = u.subschemas(node.PrefixItemsDiff)
	unrolled.ContainsDiff = u.unroll(node.ContainsDiff)
	unrolled.PatternPropertiesDiff = u.schemas(node.PatternPropertiesDiff)
	unrolled.DependentSchemasDiff = u.schemas(node.DependentSchemasDiff)
	unrolled.PropertyNamesDiff = u.unroll(node.PropertyNamesDiff)
	unrolled.UnevaluatedItemsDiff = u.unroll(node.UnevaluatedItemsDiff)
	unrolled.UnevaluatedPropertiesDiff = u.unroll(node.UnevaluatedPropertiesDiff)
	unrolled.IfDiff = u.unroll(node.IfDiff)
	unrolled.ThenDiff = u.unroll(node.ThenDiff)
	unrolled.ElseDiff = u.unroll(node.ElseDiff)
	unrolled.ContentSchemaDiff = u.unroll(node.ContentSchemaDiff)
	unrolled.DefsDiff = u.schemas(node.DefsDiff)
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
	for name, node := range diff.Modified {
		if child := u.unroll(node); child != nil {
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
	unrolled := *diff
	unrolled.Modified = ModifiedSubschemas{}
	for _, modified := range diff.Modified {
		if child := u.unroll(modified.Diff); child != nil {
			unrolled.Modified = append(unrolled.Modified, &ModifiedSubschema{Base: modified.Base, Revision: modified.Revision, Diff: child})
		}
	}
	if unrolled.Empty() {
		return nil
	}
	return &unrolled
}
