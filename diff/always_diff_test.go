package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func boolSchema(v bool) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{Always: &v}}
}

func TestSchemaDiff_Always(t *testing.T) {
	cfg := NewConfig()
	str := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	empty := &openapi3.SchemaRef{Value: &openapi3.Schema{}}

	d, err := getSchemaDiff(cfg, newState(), str, boolSchema(false))
	require.NoError(t, err)
	require.NotNil(t, d.AlwaysDiff, "a schema replaced by false is a change")
	require.Equal(t, false, d.AlwaysDiff.To)

	d, err = getSchemaDiff(cfg, newState(), boolSchema(true), boolSchema(false))
	require.NoError(t, err)
	require.NotNil(t, d.AlwaysDiff, "true to false is a change")

	d, err = getSchemaDiff(cfg, newState(), empty, boolSchema(true))
	require.NoError(t, err)
	require.NotNil(t, d.AlwaysDiff, "the document changed even though the contract did not")
}

func TestSchemaRefsValidationEquivalent_Booleans(t *testing.T) {
	cfg := NewConfig()
	empty := &openapi3.SchemaRef{Value: &openapi3.Schema{}}

	require.True(t, SchemaRefsValidationEquivalent(cfg, empty, boolSchema(true)),
		"true accepts what the empty schema accepts")
	require.True(t, SchemaRefsValidationEquivalent(cfg, boolSchema(true), boolSchema(true)))
	require.False(t, SchemaRefsValidationEquivalent(cfg, empty, boolSchema(false)),
		"false accepts nothing")
	require.False(t, SchemaRefsValidationEquivalent(cfg, boolSchema(true), boolSchema(false)))
	require.False(t, SchemaRefsValidationEquivalent(cfg, boolSchema(false), empty))
	require.True(t, SchemaRefsValidationEquivalent(cfg, boolSchema(false), boolSchema(false)))

	// Keywords beside Always are constructible only through the Go API, and
	// the boolean is authoritative: marshaling drops them, so equivalence
	// drops them too.
	mixed := boolSchema(true)
	mixed.Value.Type = &openapi3.Types{"integer"}
	require.True(t, SchemaRefsValidationEquivalent(cfg, empty, mixed))
}
