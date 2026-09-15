package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/flatten/allof"
	"github.com/stretchr/testify/require"
)

// PartnerAction is a oneOf over many branches, and each branch reaches
// PartnerAction again through nullable wrappers (an anyOf of a $ref and
// null). Every cut of that cycle makes the enclosing diffs path-dependent,
// and a diff that recomputes each path-dependent pair at every reference
// expands the fan-out into a tree of exponential size. The empty results
// must be reused across paths so that the comparison completes, and, the
// documents being identical, finds nothing. A regression hangs, which the
// package timeout reports.
func TestCircularNullableFanOut_Terminates(t *testing.T) {
	load := func() *openapi3.T {
		t.Helper()
		spec, err := openapi3.NewLoader().LoadFromFile("../data/circular-nullable-fanout.yaml")
		require.NoError(t, err)
		spec, err = allof.MergeSpec(spec)
		require.NoError(t, err)
		return spec
	}
	base, revision := load(), load()

	d, err := diff.Get(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	require.True(t, d.Empty(), "identical documents must produce an empty diff")
}
