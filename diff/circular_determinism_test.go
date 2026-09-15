package diff_test

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Token cycles through two edges, parent and sub_tokens.items. Each edge
// leads back to Token, whose diff is reported once, at the entry that reached
// it; the edges report nothing, so the output does not depend on the order
// the edges are walked in. Two fixture pairs cover the two entry points: with
// Token only reachable ref-less from components, and with a paths $ref to
// Token as well.
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
		token := d.ComponentsDiff.SchemasDiff.Modified["Token"]
		require.Contains(t, token.PropertiesDiff.Added, "index", fixture)
		require.NotContains(t, token.PropertiesDiff.Modified, "parent", "%s: the parent edge leads back to Token, which is reported once", fixture)
		subTokens := token.PropertiesDiff.Modified["sub_tokens"]
		require.NotNil(t, subTokens.DescriptionDiff, fixture)
		require.Nil(t, subTokens.ItemsDiff, "%s: the items edge leads back to Token, which is reported once", fixture)

		for range 9 {
			_, next := get(base, revision)
			require.Equal(t, string(first), string(next), "%s: identical inputs must produce identical output", fixture)
		}
	}
}
