package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/flatten/allof"
	"github.com/stretchr/testify/require"
)

// --flatten-allof merges `allOf: [$ref Node, <inline recursive overlay>]` into a
// ref-less self-referencing schema. The circular-ref guard keys on Ref strings,
// so it can't see that cycle: getSchemaDiff recursed until the stack overflowed
// (fatal error, exit 2). The in-flight schema-pair guard cuts the cycle instead,
// and real changes must still be detected.
func TestCircularAllOfOverlay_Flattened(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := loader.LoadFromFile("../data/circular-allof1.yaml")
	require.NoError(t, err)
	s2, err := loader.LoadFromFile("../data/circular-allof2.yaml")
	require.NoError(t, err)

	s1, err = allof.MergeSpec(s1)
	require.NoError(t, err)
	s2, err = allof.MergeSpec(s2)
	require.NoError(t, err)

	// used to fatal: stack overflow
	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// the description change on the recursive component is still detected
	nodeDiff := d.ComponentsDiff.SchemasDiff.Modified["Node"]
	require.NotNil(t, nodeDiff)
	require.NotNil(t, nodeDiff.DescriptionDiff)
	require.Equal(t, "recursive filter tree", nodeDiff.DescriptionDiff.From)
	require.Equal(t, "recursive filter tree, modified", nodeDiff.DescriptionDiff.To)
}
