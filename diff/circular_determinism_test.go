package diff_test

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Token cycles through two edges, parent and sub_tokens.items. Where each
// edge's walk is cut depends on the ancestors on its path, so a cut-affected
// diff holds for its own path only; reusing it for the same pair on another
// path hands one edge the other's answer, and which edge wins follows the
// map iteration order over properties, changing the output between runs
// (#1230). Each edge must carry its own one-level diff, identically on
// every run.
func TestCircularTwoEdges_Deterministic(t *testing.T) {
	get := func() (*diff.Diff, []byte) {
		t.Helper()
		loader := openapi3.NewLoader()
		s1, err := loader.LoadFromFile("../data/circular-two-edges1.yaml")
		require.NoError(t, err)
		s2, err := loader.LoadFromFile("../data/circular-two-edges2.yaml")
		require.NoError(t, err)
		d, err := diff.Get(diff.NewConfig(), s1, s2)
		require.NoError(t, err)
		paths, err := json.Marshal(d.PathsDiff)
		require.NoError(t, err)
		components, err := json.Marshal(d.ComponentsDiff)
		require.NoError(t, err)
		return d, append(paths, components...)
	}

	d, first := get()
	edges := d.ComponentsDiff.SchemasDiff.Modified["Token"].PropertiesDiff.Modified
	for _, edge := range []string{"parent", "sub_tokens"} {
		require.NotNil(t, edges[edge], "the %s cycle edge must carry its own diff", edge)
	}
	require.Contains(t, edges["parent"].PropertiesDiff.Added, "index")
	require.Contains(t, edges["sub_tokens"].ItemsDiff.PropertiesDiff.Added, "index")

	for range 9 {
		_, next := get()
		require.Equal(t, string(first), string(next), "identical inputs must produce identical output")
	}
}
