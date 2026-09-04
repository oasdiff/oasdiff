package diff

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// setBound sets the schema field whose json tag is the keyword to v. Every
// bound is numeric (a plain uint64 or a pointer to a number), so the value
// is applied by the field's kind and no per-keyword setter is needed.
func setBound(t *testing.T, s *openapi3.Schema, keyword string, v uint64) {
	t.Helper()
	typ := reflect.TypeFor[openapi3.Schema]()
	for i := range typ.NumField() {
		if name, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ","); name != keyword {
			continue
		}
		fv := reflect.ValueOf(s).Elem().Field(i)
		if fv.Kind() == reflect.Pointer {
			p := reflect.New(fv.Type().Elem())
			fv.Set(p)
			fv = p.Elem()
		}
		switch fv.Kind() {
		case reflect.Uint64:
			fv.SetUint(v)
		case reflect.Float64:
			fv.SetFloat(float64(v))
		default:
			t.Fatalf("%s: unexpected kind %s", keyword, fv.Kind())
		}
		return
	}
	t.Fatalf("no openapi3.Schema field with json tag %q", keyword)
}

func boundSchema(t *testing.T, keyword string, value uint64) *openapi3.SchemaRef {
	t.Helper()
	s := &openapi3.Schema{}
	if value != 0 {
		setBound(t, s, keyword, value)
	}
	return &openapi3.SchemaRef{Value: s}
}

// Each SchemaBounds row matches its getter's absence encoding: going from
// absent to a value classifies as WasSet, the reverse as WasUnset, and a change
// between two values as neither. A getter that changes how it encodes
// absence fails here.
func TestSchemaBounds(t *testing.T) {
	cfg := NewConfig()
	for _, bound := range SchemaBounds {
		absent := boundSchema(t, bound.Keyword, 0)
		low := boundSchema(t, bound.Keyword, 4)
		high := boundSchema(t, bound.Keyword, 8)

		set, err := getSchemaDiff(cfg, newState(), absent, low)
		require.NoError(t, err)
		value, ok := bound.WasSet(set)
		require.True(t, ok, "%s: absent to value must classify as WasSet", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.WasUnset(set)
		require.False(t, ok, bound.Keyword)

		unset, err := getSchemaDiff(cfg, newState(), low, absent)
		require.NoError(t, err)
		value, ok = bound.WasUnset(unset)
		require.True(t, ok, "%s: value to absent must classify as WasUnset", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.WasSet(unset)
		require.False(t, ok, bound.Keyword)

		changed, err := getSchemaDiff(cfg, newState(), low, high)
		require.NoError(t, err)
		_, ok = bound.WasSet(changed)
		require.False(t, ok, "%s: value to value is not WasSet", bound.Keyword)
		_, ok = bound.WasUnset(changed)
		require.False(t, ok, "%s: value to value is not WasUnset", bound.Keyword)
	}
}

// Fields whose type marks them as bounds but which SchemaBounds deliberately
// omits, with the reason.
var schemaBoundsWaived = map[string]string{
	"exclusiveMinimum": "boolean in OpenAPI 3.0 and number in 3.1; needs its own value kind",
	"exclusiveMaximum": "boolean in OpenAPI 3.0 and number in 3.1; needs its own value kind",
}

// Every openapi3.Schema field of a bound-like type (uint64, *uint64,
// *float64, ExclusiveBound) is either a SchemaBounds row or waived above
// with a reason, so a bound kin adds fails here instead of going unlisted.
func TestSchemaBoundsComplete(t *testing.T) {
	listed := map[string]bool{}
	for _, bound := range SchemaBounds {
		listed[bound.Keyword] = true
	}
	boundTypes := []reflect.Type{
		reflect.TypeFor[uint64](),
		reflect.TypeFor[*uint64](),
		reflect.TypeFor[*float64](),
		reflect.TypeFor[openapi3.ExclusiveBound](),
	}

	typ := reflect.TypeFor[openapi3.Schema]()
	for field := range typ.Fields() {
		if !slices.Contains(boundTypes, field.Type) {
			continue
		}
		keyword, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if listed[keyword] && schemaBoundsWaived[keyword] != "" {
			t.Errorf("stale waiver: %q is listed in SchemaBounds; remove the entry", keyword)
			continue
		}
		if !listed[keyword] && schemaBoundsWaived[keyword] == "" {
			t.Errorf("openapi3.Schema.%s (%s) is a bound the diff does not list\n  add it to SchemaBounds or waive it with a reason", field.Name, keyword)
		}
	}
	for keyword := range schemaBoundsWaived {
		if _, ok := typ.FieldByNameFunc(func(name string) bool {
			f, _ := typ.FieldByName(name)
			tag, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			return tag == keyword
		}); !ok {
			t.Errorf("stale waiver: no openapi3.Schema field with json tag %q; remove the entry", keyword)
		}
	}
}
