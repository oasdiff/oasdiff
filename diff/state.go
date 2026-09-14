package diff

import "github.com/getkin/kin-openapi/openapi3"

// valuePair identifies a comparison of two schema values, independent of the
// SchemaRef wrappers it arrived through.
type valuePair struct {
	value1 *openapi3.Schema
	value2 *openapi3.Schema
}

type state struct {
	// cache holds diffs that are a function of their pair alone: their
	// computation consulted no schema pair further up the stack.
	cache schemaDiffCache

	// inFlight maps each pair being diffed on the current stack to its
	// stack depth.
	inFlight map[schemaPair]int

	// frameSeq records, per stack depth, the push sequence number of the
	// frame currently at that depth. Frames pop in stack order, so a frame
	// is unchanged exactly when every frame below it is unchanged, and one
	// comparison against a recorded sequence number verifies a whole stack
	// prefix.
	frameSeq []int
	seq      int

	// scoped holds diffs whose computation cut into pairs further up the
	// stack: each stands in for computations in progress there, so it is
	// valid only while those very frames are still in flight (see
	// scopedEntry).
	scoped map[schemaPair]scopedEntry

	// cutTargets collects the stack depths that cycle cuts have targeted
	// since the current frame began; nil while none has.
	cutTargets []int

	// equivalenceInFlight holds the schema value pairs whose wrapping
	// recognition or validation-equivalence comparison is in progress in
	// this diff run or in any equivalence comparison nested inside it. An
	// equivalence comparison is itself a schema diff in a state of its own,
	// where the in-flight pair guard cannot see the outer traversal, so
	// this set is shared across the nested states (newNestedState): a
	// comparison that reaches itself again through a cyclic schema declines
	// instead of recursing forever.
	equivalenceInFlight map[valuePair]struct{}
}

// scopedEntry records that a pair's diff was empty when computed while the
// frames at depths deps were in flight and cut into: the empty answer can be
// reused only while those very frames are still in flight, verified by
// comparing the deepest one's push sequence (unchanged there means unchanged
// below, by stack order).
type scopedEntry struct {
	deps      []int
	maxDep    int
	maxDepSeq int
}

func newState() *state {
	return &state{
		cache:               schemaDiffCache{},
		inFlight:            map[schemaPair]int{},
		scoped:              map[schemaPair]scopedEntry{},
		equivalenceInFlight: map[valuePair]struct{}{},
	}
}

// newNestedState returns a state for a diff run nested inside another (an
// equivalence comparison): its own cache and cycle detection, so the outer
// traversal does not affect its results, but the outer run's equivalence
// re-entry set.
func newNestedState(outer *state) *state {
	nested := newState()
	nested.equivalenceInFlight = outer.equivalenceInFlight
	return nested
}
