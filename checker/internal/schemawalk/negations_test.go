package schemawalk

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func modifiedProperties(properties map[string]*diff.SchemaDiff) *diff.SchemasDiff {
	return &diff.SchemasDiff{Modified: properties}
}

func TestNegations(t *testing.T) {
	root := &diff.SchemaDiff{
		PropertiesDiff: modifiedProperties(map[string]*diff.SchemaDiff{
			"a": {
				NotDiff: &diff.SchemaDiff{
					PropertiesDiff: modifiedProperties(map[string]*diff.SchemaDiff{
						"b": {NotDiff: &diff.SchemaDiff{}},
					}),
				},
			},
		}),
	}

	negations := NewNegations(root)
	require.False(t, negations.At(""))
	require.False(t, negations.At("a"))
	require.True(t, negations.At("a/not"))
	require.True(t, negations.At("a/not/b"))
	require.False(t, negations.At("a/not/b/not"))
	// a path the walk did not reach takes the answer of the schema above it
	require.True(t, negations.At("a/not/added"))
	require.False(t, negations.At("added"))
}

// A property named "not" has the path of the `not` keyword beside it, so the
// path is not treated as negated.
func TestNegations_PropertyNamedNot(t *testing.T) {
	root := &diff.SchemaDiff{
		PropertiesDiff: modifiedProperties(map[string]*diff.SchemaDiff{
			"a": {
				NotDiff: &diff.SchemaDiff{},
				PropertiesDiff: modifiedProperties(map[string]*diff.SchemaDiff{
					"not": {},
				}),
			},
		}),
	}

	require.False(t, NewNegations(root).At("a/not"))
}
