package diff_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestDuplicateEndpointError_PathParamNamesDiffer(t *testing.T) {
	require.True(t, (&diff.DuplicateEndpointError{Path1: "/pet/{a}", Path2: "/pet/{b}"}).PathParamNamesDiffer())
	require.False(t, (&diff.DuplicateEndpointError{Path1: "/pets", Path2: "/pets"}).PathParamNamesDiffer())
}
