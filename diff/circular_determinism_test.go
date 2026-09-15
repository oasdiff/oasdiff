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

// S has two inline anyOf branches, listed in the opposite order in the
// revision, and one of them refers to T, whose property refers back to S.
// Reordering the branches is not a change, at S's own component entry and
// below T alike, in whichever order the components are walked.
func TestCircularSchema_ReorderedInlineBranches(t *testing.T) {
	doc := func() *openapi3.T {
		s := &openapi3.Schema{Type: &openapi3.Types{"object"}}
		tt := &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{
			"s": &openapi3.SchemaRef{Ref: "#/components/schemas/S", Value: s},
		}}
		a := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{
			"t": &openapi3.SchemaRef{Ref: "#/components/schemas/T", Value: tt},
		}}}
		b := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
		s.AnyOf = openapi3.SchemaRefs{a, b}
		return &openapi3.T{
			OpenAPI: "3.0.3",
			Info:    &openapi3.Info{Title: "t", Version: "1"},
			Paths:   openapi3.NewPaths(),
			Components: &openapi3.Components{Schemas: openapi3.Schemas{
				"S": &openapi3.SchemaRef{Value: s},
				"T": &openapi3.SchemaRef{Value: tt},
			}},
		}
	}
	base, revision := doc(), doc()
	s := revision.Components.Schemas["S"].Value
	s.AnyOf = openapi3.SchemaRefs{s.AnyOf[1], s.AnyOf[0]}

	for range 20 {
		d, err := diff.Get(diff.NewConfig(), base, revision)
		require.NoError(t, err)
		require.True(t, d.Empty(), "reordering anyOf branches is not a change")
	}
}

// Node's only anyOf branch is an inline array of Node, and the revision adds
// a property to Node. The branch is the same on both sides, so the added
// property is reported once, on Node, and the anyOf reports no branch added
// or deleted: the branch is matched by comparing it base against revision,
// where Node above it is cut, the same way as everything else on the path.
func TestCircularSchema_InlineBranchBackToSelf(t *testing.T) {
	doc := func(extra bool) *openapi3.T {
		node := &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{
			"p": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}}
		if extra {
			node.Properties["q"] = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
		}
		node.AnyOf = openapi3.SchemaRefs{{Value: &openapi3.Schema{
			Type:  &openapi3.Types{"array"},
			Items: &openapi3.SchemaRef{Ref: "#/components/schemas/Node", Value: node},
		}}}
		return &openapi3.T{
			OpenAPI:    "3.0.3",
			Info:       &openapi3.Info{Title: "t", Version: "1"},
			Paths:      openapi3.NewPaths(),
			Components: &openapi3.Components{Schemas: openapi3.Schemas{"Node": &openapi3.SchemaRef{Value: node}}},
		}
	}

	d, err := diff.Get(diff.NewConfig(), doc(false), doc(true))
	require.NoError(t, err)
	node := d.ComponentsDiff.SchemasDiff.Modified["Node"]
	require.Equal(t, []string{"q"}, node.PropertiesDiff.Added)
	require.Nil(t, node.AnyOfDiff, "the inline branch is unchanged")
}
