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
	sharedDiff := properties(map[string]*diff.SchemaDiff{"inner": {}})
	sharedDiff.RevisionComponent = "Shared"
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": sharedDiff, "b": sharedDiff}))

	shared, ok := references.At("a", 10)
	require.True(t, ok)
	require.Equal(t, schemawalk.Shared{Name: "Shared", Paths: []string{"a", "b"}, Count: 2}, shared)

	shared, ok = references.At("a/inner/x", 10)
	require.True(t, ok)
	require.Equal(t, []string{"a/inner/x", "b/inner/x"}, shared.Paths)
}

// A property whose name extends a shared one's is not below it.
func TestReferences_AtStopsAtTheSegment(t *testing.T) {
	shared := &diff.SchemaDiff{}
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": shared, "b": shared, "ab": {}}))

	_, ok := references.At("ab", 10)
	require.False(t, ok)
}

// A change in a shared schema inside another shared schema is at every
// combination of their references, and the name is the inner schema's.
func TestReferences_AtNested(t *testing.T) {
	inner := &diff.SchemaDiff{}
	inner.RevisionComponent = "Inner"
	outer := properties(map[string]*diff.SchemaDiff{"x": inner, "y": inner})
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": outer, "b": outer}))

	shared, ok := references.At("a/x/z", 10)
	require.True(t, ok)
	require.Equal(t, schemawalk.Shared{Name: "Inner", Paths: []string{"a/x/z", "b/x/z", "a/y/z", "b/y/z"}, Count: 4}, shared)
}

// Listing stops at the limit; the count still covers every path.
func TestReferences_AtLimit(t *testing.T) {
	inner := &diff.SchemaDiff{}
	outer := properties(map[string]*diff.SchemaDiff{"x": inner, "y": inner})
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": outer, "b": outer}))

	shared, ok := references.At("a/x/z", 2)
	require.True(t, ok)
	require.Equal(t, []string{"a/x/z", "b/x/z"}, shared.Paths)
	require.Equal(t, 4, shared.Count)
}

// A schema that contains itself is at unboundedly many paths. Listing them
// still ends at the limit.
func TestReferences_AtCycle(t *testing.T) {
	node := &diff.SchemaDiff{PropertiesDiff: &diff.SchemasDiff{Modified: diff.ModifiedSchemasMap{"name": {}}}}
	node.PropertiesDiff.Modified["child"] = node
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"root": node}))

	shared, ok := references.At("root/name", 3)
	require.True(t, ok)
	require.True(t, shared.Cyclic)
	require.Zero(t, shared.Count)
	require.Equal(t, []string{"root/name", "root/child/name", "root/child/child/name"}, shared.Paths)
}

// A shared schema with no name of its own takes the name of the innermost
// named schema the walk passed through on the way to it.
func TestReferences_InlineSharedSchemaIsNamedAfterItsComponent(t *testing.T) {
	inline := &diff.SchemaDiff{}
	component := properties(map[string]*diff.SchemaDiff{"x": inline, "y": inline})
	component.RevisionComponent = "Component"
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"c": component}))

	shared, ok := references.At("c/x", 10)
	require.True(t, ok)
	require.Equal(t, "Component", shared.Name)
}

// A schema reached once is not shared.
func TestReferences_SchemaReachedOnce(t *testing.T) {
	references := schemawalk.NewReferences(properties(map[string]*diff.SchemaDiff{"a": {}, "b": {}}))

	_, ok := references.At("a", 10)
	require.False(t, ok)
}
