package checker

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// The comparison is exact rational arithmetic on the decimals the values
// round-trip to, so decimal literals divide the way their author meant
// (0.3 over 0.1 is 3) and a ratio that is merely close to an integer is not
// treated as one: reporting a multipleOf change as a generalization claims
// that every previously valid value remains valid, which a near-integer
// ratio does not deliver.
func TestIsIntegerMultiple(t *testing.T) {
	require.True(t, isIntegerMultiple(6, 3))
	require.True(t, isIntegerMultiple(0.3, 0.1))
	require.True(t, isIntegerMultiple(5, 2.5))
	require.True(t, isIntegerMultiple(0.000006, 0.000002))

	require.False(t, isIntegerMultiple(7, 3))
	require.False(t, isIntegerMultiple(0.1, 0.3))
	require.False(t, isIntegerMultiple(1.000000001, 1),
		"a ratio within float tolerance of an integer is not an integer: 1.000000001 itself would be rejected under multipleOf 1")

	require.False(t, isIntegerMultiple(1, 0))
	require.False(t, isIntegerMultiple(math.NaN(), 1))
	require.False(t, isIntegerMultiple(math.Inf(1), 1))
}
