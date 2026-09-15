package diff

import (
	"math"

	"github.com/getkin/kin-openapi/openapi3"
)

// valuePair identifies a comparison of two schema values, independent of the
// SchemaRef wrappers it arrived through.
type valuePair struct {
	value1 *openapi3.Schema
	value2 *openapi3.Schema
}

type state struct {
	// cache holds diffs that are a function of their pair alone: no cut in
	// their computation reached a frame above them.
	cache schemaDiffCache

	// inFlight maps each pair being diffed on the current stack to its
	// stack depth.
	inFlight map[valuePair]int

	// minCutTarget is the shallowest stack depth that a cycle cut has
	// targeted since the current frame began; math.MaxInt while none has.
	minCutTarget int

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

func newState() *state {
	return &state{
		cache:               schemaDiffCache{},
		inFlight:            map[valuePair]int{},
		minCutTarget:        math.MaxInt,
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
