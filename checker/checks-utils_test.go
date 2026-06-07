package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// #916: the property-diff traversal helpers must descend into the `not`
// sub-schema, like they do for if/then/else and the other sub-schema fields.
// Before the fix, a change inside `not` was populated in NotDiff but never
// reached the visitor, so every check built on these helpers silently missed it.

func TestCheckModifiedPropertiesDiff_TraversesNot(t *testing.T) {
	schemaDiff := &diff.SchemaDiff{
		NotDiff: &diff.SchemaDiff{
			PropertiesDiff: &diff.SchemasDiff{
				Modified: diff.ModifiedSchemasMap{"legacy_field": {}},
			},
		},
	}

	var visited []string
	checker.CheckModifiedPropertiesDiff(schemaDiff, func(propertyPath, propertyName string, _ *diff.SchemaDiff, _ *diff.SchemaDiff) {
		if propertyName != "" {
			visited = append(visited, propertyPath+"/"+propertyName)
		}
	})

	require.Contains(t, visited, "/not/legacy_field")
}

func TestCheckAddedPropertiesDiff_TraversesNot(t *testing.T) {
	schemaDiff := &diff.SchemaDiff{
		NotDiff: &diff.SchemaDiff{
			Revision: &openapi3.Schema{
				Properties: openapi3.Schemas{"new_field": &openapi3.SchemaRef{Value: &openapi3.Schema{}}},
			},
			PropertiesDiff: &diff.SchemasDiff{Added: []string{"new_field"}},
		},
	}

	var added []string
	checker.CheckAddedPropertiesDiff(schemaDiff, func(propertyPath, propertyName string, _ *openapi3.Schema, _ *diff.SchemaDiff) {
		added = append(added, propertyPath+"/"+propertyName)
	})

	require.Contains(t, added, "/not/new_field")
}

func TestCheckDeletedPropertiesDiff_TraversesNot(t *testing.T) {
	schemaDiff := &diff.SchemaDiff{
		NotDiff: &diff.SchemaDiff{
			Base: &openapi3.Schema{
				Properties: openapi3.Schemas{"old_field": &openapi3.SchemaRef{Value: &openapi3.Schema{}}},
			},
			PropertiesDiff: &diff.SchemasDiff{Deleted: []string{"old_field"}},
		},
	}

	var deleted []string
	checker.CheckDeletedPropertiesDiff(schemaDiff, func(propertyPath, propertyName string, _ *openapi3.Schema, _ *diff.SchemaDiff) {
		deleted = append(deleted, propertyPath+"/"+propertyName)
	})

	require.Contains(t, deleted, "/not/old_field")
}
