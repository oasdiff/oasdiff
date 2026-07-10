package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestGetNullableWrappingDiff(t *testing.T) {
	cfg := NewConfig()
	str := &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"a", "b"}}
	null := &openapi3.Schema{Type: &openapi3.Types{"null"}}
	wrap := func(branches ...*openapi3.Schema) *openapi3.Schema {
		refs := make(openapi3.SchemaRefs, 0, len(branches))
		for _, b := range branches {
			refs = append(refs, &openapi3.SchemaRef{Value: b})
		}
		return &openapi3.Schema{OneOf: refs}
	}

	require.False(t, getNullableWrappingDiff(cfg, str, wrap(null, str)).Empty(), "the pure wrap is recognized")
	require.False(t, getNullableWrappingDiff(cfg, str, wrap(str, null)).Empty(), "branch order does not matter")

	nullableStr := &openapi3.Schema{Type: &openapi3.Types{"string", "null"}}
	require.True(t, getNullableWrappingDiff(cfg, nullableStr, wrap(null, nullableStr)).Empty(),
		"wrapping an already-nullable schema is not a pure widening under oneOf")

	other := &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"a"}}
	require.True(t, getNullableWrappingDiff(cfg, str, wrap(null, other)).Empty(),
		"a non-equivalent branch is not recognized")

	require.True(t, getNullableWrappingDiff(cfg, str, wrap(null, str, str)).Empty(),
		"only a two-branch wrap is recognized")

	nullWithEnum := &openapi3.Schema{Type: &openapi3.Types{"null"}, Enum: []any{"x"}}
	require.True(t, getNullableWrappingDiff(cfg, str, wrap(nullWithEnum, str)).Empty(),
		"a constrained null branch is not a bare null")

	// The bare-ness checks are by validation equivalence to the empty schema,
	// not a field list: an unanticipated constraining keyword on the wrapper
	// declines the recognition, while annotations don't.
	constrained := wrap(null, str)
	constrained.MinProps = 1
	require.True(t, getNullableWrappingDiff(cfg, str, constrained).Empty(),
		"a wrapper with its own constraints is not bare")

	annotated := wrap(null, str)
	annotated.Description = "now nullable"
	require.False(t, getNullableWrappingDiff(cfg, str, annotated).Empty(),
		"annotations on the wrapper don't block the recognition")

	nullDescribed := &openapi3.Schema{Type: &openapi3.Types{"null"}, Description: "no value"}
	require.False(t, getNullableWrappingDiff(cfg, str, wrap(nullDescribed, str)).Empty(),
		"annotations on the null branch don't block the recognition")
}
