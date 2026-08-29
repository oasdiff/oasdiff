package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestGetPrefixItemsDiff_Positional(t *testing.T) {
	cfg := NewConfig()
	str := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	}
	integer := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}
	}
	boolean := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}}}
	}
	prefix := func(refs ...*openapi3.SchemaRef) openapi3.SchemaRefs { return refs }

	// reordering: each position now validates against a different schema, so
	// both positions are modified; the content matcher reported no change
	d, err := getPrefixItemsDiff(cfg, newState(), prefix(str(), integer()), prefix(integer(), str()))
	require.NoError(t, err)
	require.Len(t, d.Modified, 2)
	require.Empty(t, d.Added)
	require.Empty(t, d.Deleted)
	require.Equal(t, 0, d.Modified[0].Base.Index)
	require.Equal(t, 0, d.Modified[0].Revision.Index)

	// inserting at the front shifts every later position: both existing
	// positions change type and a third appears
	d, err = getPrefixItemsDiff(cfg, newState(), prefix(str(), integer()), prefix(boolean(), str(), integer()))
	require.NoError(t, err)
	require.Len(t, d.Modified, 2)
	require.Len(t, d.Added, 1)
	require.Equal(t, 2, d.Added[0].Index)

	// appending leaves the covered positions untouched
	d, err = getPrefixItemsDiff(cfg, newState(), prefix(str(), integer()), prefix(str(), integer(), boolean()))
	require.NoError(t, err)
	require.Empty(t, d.Modified)
	require.Len(t, d.Added, 1)
	require.Equal(t, 2, d.Added[0].Index)

	// truncating deletes at the index
	d, err = getPrefixItemsDiff(cfg, newState(), prefix(str(), integer()), prefix(str()))
	require.NoError(t, err)
	require.Empty(t, d.Modified)
	require.Len(t, d.Deleted, 1)
	require.Equal(t, 1, d.Deleted[0].Index)

	// identical lists report nothing
	d, err = getPrefixItemsDiff(cfg, newState(), prefix(str(), integer()), prefix(str(), integer()))
	require.NoError(t, err)
	require.Nil(t, d)
}
