package diff

type state struct {
	graph schemaGraph

	// equivalenceInFlight holds the schema value pairs whose wrapping
	// recognition or validation-equivalence comparison is in progress in
	// this diff run or in any equivalence comparison nested inside it. An
	// equivalence comparison is itself a schema diff with a graph of its
	// own, which cannot see the outer traversal, so this set is shared
	// across the nested states (newNestedState): a comparison that reaches
	// itself again through a cyclic schema declines instead of recursing
	// forever.
	equivalenceInFlight map[valuePair]struct{}
}

func newState() *state {
	return &state{
		graph:               newSchemaGraph(),
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
