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

// A change in a shared schema stays one change, at the property it was found
// at, and lists the others.
func TestAttachSharedSchemas_OneChange(t *testing.T) {
	root, shared := sharedRoot()
	change := ApiChange{Id: "change_id", Args: []any{"a", "200"}}.WithSchema(root, shared, "a")

	result := attachSharedSchemas(Changes{change})

	require.Len(t, result, 1)
	require.Equal(t, []any{"a", "200"}, result[0].GetArgs())
	require.Equal(t, &SharedSchema{Properties: []string{"a", "b"}, Count: 2}, result[0].(ApiChange).GetSharedSchema())
	require.Equal(t, SharedSchemaCommentId, result[0].(ApiChange).Comment)
}

// A check's own comment is kept.
func TestAttachSharedSchemas_KeepsComment(t *testing.T) {
	root, shared := sharedRoot()
	change := ApiChange{Id: "change_id", Comment: "own-comment"}.WithSchema(root, shared, "a")

	result := attachSharedSchemas(Changes{change})

	require.Equal(t, "own-comment", result[0].(ApiChange).Comment)
}

// A change at the root, or with no schema, is in no shared schema.
func TestAttachSharedSchemas_NotShared(t *testing.T) {
	root, _ := sharedRoot()
	atRoot := ApiChange{Id: "change_id"}.WithSchema(root, root, "")
	noSchema := ApiChange{Id: "change_id"}

	result := attachSharedSchemas(Changes{atRoot, noSchema})

	require.Nil(t, result[0].(ApiChange).GetSharedSchema())
	require.Nil(t, result[1].(ApiChange).GetSharedSchema())
}

// The detail lists the properties after the first and counts the ones not
// listed.
func TestSharedSchemaDetail(t *testing.T) {
	l := NewLocalizer("en")
	for _, tc := range []struct {
		name   string
		shared SharedSchema
		want   string
	}{
		{"all listed", SharedSchema{Name: "Id", Properties: []string{"a", "b", "c"}, Count: 3}, "(shared schema: `Id`, also at `b`, `c`)"},
		{"more", SharedSchema{Name: "Id", Properties: []string{"a", "b", "c"}, Count: 7}, "(shared schema: `Id`, also at `b`, `c` and 4 more)"},
		{"cyclic", SharedSchema{Name: "Id", Properties: []string{"a", "b", "c"}, Cyclic: true}, "(shared schema: `Id`, also at `b`, `c` and more)"},
		{"no name", SharedSchema{Properties: []string{"a", "b"}, Count: 2}, "(also at `b`)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, tc.shared.detail(l, quotedValues))
		})
	}
}
