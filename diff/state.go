package diff

import "github.com/getkin/kin-openapi/openapi3"

// valuePair identifies a comparison of two schema values, independent of the
// SchemaRef wrappers it arrived through.
type valuePair struct {
	value1 *openapi3.Schema
	value2 *openapi3.Schema
}

type state struct {
	// nodes holds the schema diff graph: one node per pair of schema values,
	// linked to the nodes of its sub-schema pairs (see getSchemaDiffNode).
	nodes map[valuePair]*SchemaDiff

	// inProgress holds the nodes whose diff is being computed on the current
	// stack; their fields are not filled yet.
	inProgress map[*SchemaDiff]struct{}

	// depth is the number of schema diffs on the current stack; zero when a
	// pair is reached from outside the schema graph.
	depth int

	// unrolled memoizes the tree of each node whose unrolling did not cut
	// into a node above it, so the tree is the same on every path (see
	// unroller).
	unrolled map[*SchemaDiff]*SchemaDiff

	// equivalenceInFlight holds the schema value pairs whose wrapping
	// recognition or validation-equivalence comparison is in progress in
	// this diff run or in any equivalence comparison nested inside it. An
	// equivalence comparison is itself a schema diff in a state of its own,
	// where the graph cannot see the outer traversal, so this set is shared
	// across the nested states (newNestedState): a comparison that reaches
	// itself again through a cyclic schema declines instead of recursing
	// forever.
	equivalenceInFlight map[valuePair]struct{}
}

func newState() *state {
	return &state{
		nodes:               map[valuePair]*SchemaDiff{},
		inProgress:          map[*SchemaDiff]struct{}{},
		unrolled:            map[*SchemaDiff]*SchemaDiff{},
		equivalenceInFlight: map[valuePair]struct{}{},
	}
}

// newNestedState returns a state for a diff run nested inside another (an
// equivalence comparison): its own graph, so the outer traversal does not
// affect its results, but the outer run's equivalence re-entry set.
func newNestedState(outer *state) *state {
	nested := newState()
	nested.equivalenceInFlight = outer.equivalenceInFlight
	return nested
}
