package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestPrefixItemsValidationEquivalent(t *testing.T) {
	cfg := NewConfig()
	str := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	}
	integer := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}
	}
	arr := func(items *openapi3.SchemaRef, prefix ...*openapi3.SchemaRef) *openapi3.Schema {
		return &openapi3.Schema{Type: &openapi3.Types{"array"}, Items: items, PrefixItems: prefix}
	}

	require.True(t, PrefixItemsValidationEquivalent(cfg, arr(str()), arr(str(), str())),
		"an entry repeating the items schema validates its position exactly as items did")
	require.False(t, PrefixItemsValidationEquivalent(cfg, arr(str()), arr(str(), integer())),
		"an entry differing from the items schema changes its position")

	// With no items schema, a position past prefixItems is unconstrained, so
	// adding an entry constrains it and removing one opens it.
	require.False(t, PrefixItemsValidationEquivalent(cfg, arr(nil), arr(nil, str())),
		"an entry added where nothing constrained the position is a change")
	require.True(t, PrefixItemsValidationEquivalent(cfg, arr(nil), arr(nil, &openapi3.SchemaRef{Value: &openapi3.Schema{}})),
		"an entry that constrains nothing leaves an unconstrained position unconstrained")

	require.True(t, PrefixItemsValidationEquivalent(cfg, arr(str(), integer()), arr(str(), integer())),
		"identical prefixItems are equivalent")
	require.False(t, PrefixItemsValidationEquivalent(cfg, arr(str(), str(), integer()), arr(str(), integer(), str())),
		"reordering entries changes the positions they validate")

	annotated := integer()
	annotated.Value.Description = "the count"
	require.True(t, PrefixItemsValidationEquivalent(cfg, arr(str(), integer()), arr(str(), annotated)),
		"an annotation-only difference does not change the contract")

	require.False(t, PrefixItemsValidationEquivalent(cfg, nil, arr(str())))
	require.False(t, PrefixItemsValidationEquivalent(cfg, arr(str()), nil))
}
