package schemawalk_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func properties(children map[string]*diff.SchemaDiff) *diff.SchemaDiff {
	return &diff.SchemaDiff{PropertiesDiff: &diff.SchemasDiff{Modified: diff.ModifiedSchemasMap(children)}}
}

// A change in a schema two properties reach, and one below it, are at both.
func TestReferences_At(t *testing.T) {
	shared := properties(map[string]*diff.SchemaDiff{"inner": {}})
	shared.RevisionComponent = "Shared"
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": shared, "b": shared}))

	name, paths, ok := references.At("a")
	require.True(t, ok)
	require.Equal(t, "Shared", name)
	require.Equal(t, []string{"a", "b"}, paths)

	_, paths, ok = references.At("a/inner/x")
	require.True(t, ok)
	require.Equal(t, []string{"a/inner/x", "b/inner/x"}, paths)
}

// A property whose name extends a shared one's is not below it.
func TestReferences_AtStopsAtTheSegment(t *testing.T) {
	shared := &diff.SchemaDiff{}
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": shared, "b": shared, "ab": {}}))

	_, _, ok := references.At("ab")
	require.False(t, ok)
}

// Of two shared schemas around a change, the inner one is the one reported.
func TestReferences_AtInnermost(t *testing.T) {
	inner := &diff.SchemaDiff{}
	outer := properties(map[string]*diff.SchemaDiff{"x": inner, "y": inner})
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": outer, "b": outer}))

	_, paths, ok := references.At("a/x/z")
	require.True(t, ok)
	require.Equal(t, []string{"a/x/z", "a/y/z"}, paths)
}

// A shared schema with no name of its own takes the name of the innermost
// named schema the walk passed through on the way to it.
func TestReferences_InlineSharedSchemaIsNamedAfterItsComponent(t *testing.T) {
	inline := &diff.SchemaDiff{}
	component := properties(map[string]*diff.SchemaDiff{"x": inline, "y": inline})
	component.RevisionComponent = "Component"
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"c": component}))

	name, _, ok := references.At("c/x")
	require.True(t, ok)
	require.Equal(t, "Component", name)
}

// A schema reached once is not shared.
func TestReferences_SchemaReachedOnce(t *testing.T) {
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": {}, "b": {}}))

	_, _, ok := references.At("a")
	require.False(t, ok)
}
