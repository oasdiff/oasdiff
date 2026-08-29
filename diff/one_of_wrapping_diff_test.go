package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestGetOneOfWrappingDiff_OriginalPreserved(t *testing.T) {
	cfg := NewConfig()
	obj := func(required []string, props map[string]*openapi3.Schema) *openapi3.Schema {
		schemas := openapi3.Schemas{}
		for name, s := range props {
			schemas[name] = &openapi3.SchemaRef{Value: s}
		}
		return &openapi3.Schema{Type: &openapi3.Types{"object"}, Required: required, Properties: schemas}
	}
	str := func() *openapi3.Schema { return &openapi3.Schema{Type: &openapi3.Types{"string"}} }
	wrap := func(branches ...*openapi3.Schema) *openapi3.Schema {
		refs := make(openapi3.SchemaRefs, 0, len(branches))
		for _, b := range branches {
			refs = append(refs, &openapi3.SchemaRef{Value: b})
		}
		return &openapi3.Schema{OneOf: refs}
	}

	base := obj([]string{"foo"}, map[string]*openapi3.Schema{"foo": str(), "bar": str()})
	other := obj([]string{"ref"}, map[string]*openapi3.Schema{"ref": str()})

	same := obj([]string{"foo"}, map[string]*openapi3.Schema{"foo": str(), "bar": str()})
	require.True(t, getOneOfWrappingDiff(cfg, base, wrap(same, other)).OriginalPreserved,
		"an alternative with the base's validation contract preserves the original")
	require.True(t, getOneOfWrappingDiff(cfg, base, wrap(other, same)).OriginalPreserved,
		"branch order does not matter")

	relaxed := obj(nil, map[string]*openapi3.Schema{"foo": str(), "bar": str()})
	require.False(t, getOneOfWrappingDiff(cfg, base, wrap(relaxed, other)).OriginalPreserved,
		"an alternative that drops a required property is not the original")

	narrowed := obj([]string{"foo"}, map[string]*openapi3.Schema{
		"foo": {Type: &openapi3.Types{"string"}, MinLength: 1},
		"bar": str(),
	})
	require.False(t, getOneOfWrappingDiff(cfg, base, wrap(narrowed, other)).OriginalPreserved,
		"an alternative that constrains a property is not the original")

	// Equivalence is by validation contract, so an annotation on an otherwise
	// identical alternative still counts as the original.
	annotated := obj([]string{"foo"}, map[string]*openapi3.Schema{"foo": str(), "bar": str()})
	annotated.Description = "the original, documented"
	require.True(t, getOneOfWrappingDiff(cfg, base, wrap(annotated, other)).OriginalPreserved,
		"an annotation-only difference does not change the contract")
}
