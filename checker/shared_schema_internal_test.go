package checker

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// sharedRoot is a payload whose properties a and b reference one schema.
func sharedRoot() (root *diff.SchemaDiff, shared *diff.SchemaDiff) {
	shared = &diff.SchemaDiff{}
	root = &diff.SchemaDiff{PropertiesDiff: &diff.SchemasDiff{Modified: diff.ModifiedSchemasMap{"a": shared, "b": shared}}}
	return root, shared
}

// The copy for b names b where the original named a.
func TestExpandSharedSchemas_CopyPerReference(t *testing.T) {
	root, shared := sharedRoot()
	change := ApiChange{Id: "change_id", Args: []any{"a", "200"}}.WithSchema(root, shared, "a")

	result := expandSharedSchemas(Changes{change})

	require.Len(t, result, 2)
	require.Equal(t, []any{"a", "200"}, result[0].GetArgs())
	require.Equal(t, &SharedSchema{Properties: []string{"a", "b"}}, result[0].(ApiChange).GetSharedSchema())
	require.Equal(t, []any{"b", "200"}, result[1].GetArgs())
	require.Equal(t, &SharedSchema{Properties: []string{"b", "a"}}, result[1].(ApiChange).GetSharedSchema())
}

// A message that names no property could not be told apart from its copies,
// so it is reported once, listing the properties.
func TestExpandSharedSchemas_MessageNamingNoProperty(t *testing.T) {
	root, shared := sharedRoot()
	change := ApiChange{Id: "change_id", Args: []any{"X-Header"}}.WithSchema(root, shared, "a")

	result := expandSharedSchemas(Changes{change})

	require.Len(t, result, 1)
	require.Equal(t, &SharedSchema{Properties: []string{"a", "b"}}, result[0].(ApiChange).GetSharedSchema())
}

// The changes returned do not hold on to the diff, shared or not.
func TestExpandSharedSchemas_ForgetsWhereChangesWereComputed(t *testing.T) {
	root, shared := sharedRoot()

	result := expandSharedSchemas(Changes{
		ApiChange{Id: "change_id", Args: []any{"a"}}.WithSchema(root, shared, "a"),
		ApiChange{Id: "change_id"}.WithSchema(root, root, ""),
	})

	for _, change := range result {
		apiChange := change.(ApiChange)
		require.Nil(t, apiChange.root)
		require.Nil(t, apiChange.schema)
		require.Empty(t, apiChange.propertyPath)
	}
}
