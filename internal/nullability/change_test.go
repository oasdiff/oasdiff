package nullability_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/internal/nullability"
	"github.com/oasdiff/oasdiff/internal/oasversion"
	"github.com/stretchr/testify/require"
)

func TestChange31(t *testing.T) {
	null, object, nullableObject := typed("null"), typed("object"), typed("object", "null")
	one := &openapi3.Schema{OneOf: refs(null, object)}
	two := &openapi3.Schema{OneOf: refs(null, nullableObject)}

	change := func(base, revision *openapi3.Schema) (nullability.Transition, bool) {
		return nullability.Change(base, revision, oasversion.Dialects{Base: oasversion.OpenAPI31, Revision: oasversion.OpenAPI31})
	}
	for name, tc := range map[string]struct {
		base, revision *openapi3.Schema
		want           nullability.Transition
		decided        bool
	}{
		"a second branch became nullable":   {one, two, nullability.BecameNotNullable, true},
		"one of two stopped being nullable": {two, one, nullability.BecameNullable, true},
		"nothing changed":                   {one, one, nullability.Unchanged, true},
		"a typed schema was wrapped":        {object, one, nullability.BecameNullable, true},
		"an untyped schema was wrapped":     {&openapi3.Schema{}, &openapi3.Schema{OneOf: refs(null, &openapi3.Schema{})}, nullability.BecameNotNullable, true},
		"the wrapper was removed":           {one, object, nullability.BecameNotNullable, true},
		"a branch gains null beside nullable": {
			&openapi3.Schema{Nullable: true, OneOf: refs(typed("string"), typed("array"))},
			&openapi3.Schema{Nullable: true, OneOf: refs(&openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}, typed("array"))},
			nullability.Unchanged, true,
		},
		"one side is not decided": {&openapi3.Schema{AllOf: refs(object)}, one, nullability.Unchanged, false},
	} {
		t.Run(name, func(t *testing.T) {
			got, decided := change(tc.base, tc.revision)
			require.Equal(t, tc.want, got)
			require.Equal(t, tc.decided, decided)
		})
	}
}

// Each side is read in its own document's dialect, so a schema that moved
// from the OpenAPI 3.0 way of saying nullable to the 3.1 way did not change.
func TestChange_AcrossDialects(t *testing.T) {
	migration := oasversion.Dialects{Base: oasversion.OpenAPI30, Revision: oasversion.OpenAPI31}
	nullableString30 := &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}
	untyped := &openapi3.Schema{}

	got, decided := nullability.Change(nullableString30, typed("string", "null"), migration)
	require.True(t, decided)
	require.Equal(t, nullability.Unchanged, got)

	// the same schema with no type accepts null only once it is read as 3.1
	got, decided = nullability.Change(untyped, untyped, migration)
	require.True(t, decided)
	require.Equal(t, nullability.BecameNullable, got)
}
