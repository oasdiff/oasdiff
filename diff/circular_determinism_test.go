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
// path hands one edge the other's answer (#1230). Two fixture pairs cover
// the two ways that goes wrong: with Token only reachable ref-less from
// components, the two edges race on the map iteration order over properties
// and the output changes between runs; with a paths $ref to Token as well,
// the paths entry computes the edge pairs under its own visited set first,
// and reusing those cuts truncates both edges of the components entry.
// Each edge must carry its own one-level diff, identically on every run.
func TestCircularTwoEdges_Deterministic(t *testing.T) {
	get := func(base, revision string) (*diff.Diff, []byte) {
		t.Helper()
		loader := openapi3.NewLoader()
		s1, err := loader.LoadFromFile(base)
		require.NoError(t, err)
		s2, err := loader.LoadFromFile(revision)
		require.NoError(t, err)
		d, err := diff.Get(diff.NewConfig(), s1, s2)
		require.NoError(t, err)
		paths, err := json.Marshal(d.PathsDiff)
		require.NoError(t, err)
		components, err := json.Marshal(d.ComponentsDiff)
		require.NoError(t, err)
		return d, append(paths, components...)
	}

	for _, fixture := range []string{"../data/circular-two-edges", "../data/circular-two-edges-components"} {
		base, revision := fixture+"1.yaml", fixture+"2.yaml"

		d, first := get(base, revision)
		edges := d.ComponentsDiff.SchemasDiff.Modified["Token"].PropertiesDiff.Modified
		for _, edge := range []string{"parent", "sub_tokens"} {
			require.NotNil(t, edges[edge], "%s: the %s cycle edge must carry its own diff", fixture, edge)
		}
		require.Contains(t, edges["parent"].PropertiesDiff.Added, "index", fixture)
		require.Contains(t, edges["sub_tokens"].ItemsDiff.PropertiesDiff.Added, "index", fixture)

		for range 9 {
			_, next := get(base, revision)
			require.Equal(t, string(first), string(next), "%s: identical inputs must produce identical output", fixture)
		}
	}
}
