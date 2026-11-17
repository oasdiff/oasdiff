package nullability_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/internal/nullability"
	"github.com/stretchr/testify/require"
)

func typed(types ...string) *openapi3.Schema {
	t := openapi3.Types(types)
	return &openapi3.Schema{Type: &t}
}

func refs(schemas ...*openapi3.Schema) openapi3.SchemaRefs {
	result := make(openapi3.SchemaRefs, len(schemas))
	for i, schema := range schemas {
		result[i] = &openapi3.SchemaRef{Value: schema}
	}
	return result
}

func TestTypeAccepts(t *testing.T) {
	require.True(t, nullability.TypeAccepts(typed("string", "null")))
	require.True(t, nullability.TypeAccepts(&openapi3.Schema{Nullable: true}))
	require.False(t, nullability.TypeAccepts(typed("string")))
	// no type keyword names no type, null included
	require.False(t, nullability.TypeAccepts(&openapi3.Schema{}))
}

func TestOwnKeywordsAccept(t *testing.T) {
	for name, tc := range map[string]struct {
		schema *openapi3.Schema
		want   bool
	}{
		"no type accepts every type":      {&openapi3.Schema{}, true},
		"a type without null":             {typed("string"), false},
		"a type with null":                {typed("string", "null"), true},
		"nullable with a type":            {&openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}, true},
		"an enum without null":            {&openapi3.Schema{Enum: []any{"a"}}, false},
		"an enum with null":               {&openapi3.Schema{Enum: []any{"a", nil}}, true},
		"an enum overrides a null type":   {&openapi3.Schema{Type: &openapi3.Types{"string", "null"}, Enum: []any{"a"}}, false},
		"a const other than null":         {&openapi3.Schema{Const: "a"}, false},
		"a type without null and an enum": {&openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{nil}}, false},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, nullability.OwnKeywordsAccept(tc.schema))
		})
	}
}

func TestComposedAccepts(t *testing.T) {
	null, object, nullableObject := typed("null"), typed("object"), typed("object", "null")

	for name, tc := range map[string]struct {
		schema  *openapi3.Schema
		accepts bool
		decided bool
	}{
		"nil":                                 {nil, false, false},
		"no composition":                      {object, false, false},
		"oneOf with one nullable branch":      {&openapi3.Schema{OneOf: refs(null, object)}, true, true},
		"oneOf with two nullable branches":    {&openapi3.Schema{OneOf: refs(null, nullableObject)}, false, true},
		"oneOf with no nullable branch":       {&openapi3.Schema{OneOf: refs(object, typed("string"))}, false, true},
		"oneOf with an untyped branch":        {&openapi3.Schema{OneOf: refs(null, &openapi3.Schema{})}, false, true},
		"anyOf with two nullable branches":    {&openapi3.Schema{AnyOf: refs(null, nullableObject)}, true, true},
		"anyOf with no nullable branch":       {&openapi3.Schema{AnyOf: refs(object)}, false, true},
		"its own type rejects null":           {&openapi3.Schema{Type: &openapi3.Types{"object"}, OneOf: refs(null, object)}, false, true},
		"an allOf alongside":                  {&openapi3.Schema{OneOf: refs(null, object), AllOf: refs(object)}, false, false},
		"a not alongside":                     {&openapi3.Schema{OneOf: refs(null, object), Not: &openapi3.SchemaRef{Value: object}}, false, false},
		"an if alongside":                     {&openapi3.Schema{OneOf: refs(null, object), If: &openapi3.SchemaRef{Value: object}}, false, false},
		"a branch that composes further":      {&openapi3.Schema{OneOf: refs(null, &openapi3.Schema{AnyOf: refs(object)})}, false, false},
		"a branch with no schema":             {&openapi3.Schema{OneOf: openapi3.SchemaRefs{{}}}, false, false},
		"oneOf and anyOf, both accepting":     {&openapi3.Schema{OneOf: refs(null, object), AnyOf: refs(null)}, true, true},
		"oneOf accepting, anyOf not":          {&openapi3.Schema{OneOf: refs(null, object), AnyOf: refs(object)}, false, true},
		"an undecided anyOf beside the oneOf": {&openapi3.Schema{OneOf: refs(null, object), AnyOf: openapi3.SchemaRefs{nil}}, false, false},
	} {
		t.Run(name, func(t *testing.T) {
			accepts, decided := nullability.ComposedAccepts(tc.schema)
			require.Equal(t, tc.accepts, accepts)
			require.Equal(t, tc.decided, decided)
		})
	}
}

func TestAccepts(t *testing.T) {
	null, object := typed("null"), typed("object")

	for name, tc := range map[string]struct {
		schema  *openapi3.Schema
		accepts bool
		decided bool
	}{
		"nil":                        {nil, false, false},
		"a type without null":        {object, false, true},
		"no type":                    {&openapi3.Schema{}, true, true},
		"a oneOf decides":            {&openapi3.Schema{OneOf: refs(null, object)}, true, true},
		"an allOf is not decided":    {&openapi3.Schema{AllOf: refs(object)}, false, false},
		"a not is not decided":       {&openapi3.Schema{Not: &openapi3.SchemaRef{Value: object}}, false, false},
		"an if is not decided":       {&openapi3.Schema{If: &openapi3.SchemaRef{Value: object}}, false, false},
		"nullable beside a oneOf":    {&openapi3.Schema{Nullable: true, OneOf: refs(object, typed("string"))}, true, true},
		"nullable beside an allOf":   {&openapi3.Schema{Nullable: true, OneOf: refs(object), AllOf: refs(object)}, false, false},
		"a oneOf with no null in it": {&openapi3.Schema{OneOf: refs(object, typed("string"))}, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			accepts, decided := nullability.Accepts(tc.schema)
			require.Equal(t, tc.accepts, accepts)
			require.Equal(t, tc.decided, decided)
		})
	}
}

// A const written as null in a document accepts null, and one with another
// value does not.
func TestOwnKeywordsAccept_ConstInDocument(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: test
  version: 1.0.0
paths: {}
components:
  schemas:
    Nothing:
      const: null
    Home:
      const: home
`))
	require.NoError(t, err)

	schemas := doc.Components.Schemas
	require.True(t, nullability.OwnKeywordsAccept(schemas["Nothing"].Value))
	require.False(t, nullability.OwnKeywordsAccept(schemas["Home"].Value))
}
