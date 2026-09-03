package diff

import (
	"reflect"
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
// absent to a value classifies as Set, the reverse as Unset, and a change
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
		value, ok := bound.Set(set)
		require.True(t, ok, "%s: absent to value must classify as Set", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.Unset(set)
		require.False(t, ok, bound.Keyword)

		unset, err := getSchemaDiff(cfg, newState(), low, absent)
		require.NoError(t, err)
		value, ok = bound.Unset(unset)
		require.True(t, ok, "%s: value to absent must classify as Unset", bound.Keyword)
		require.NotNil(t, value, bound.Keyword)
		_, ok = bound.Set(unset)
		require.False(t, ok, bound.Keyword)

		changed, err := getSchemaDiff(cfg, newState(), low, high)
		require.NoError(t, err)
		_, ok = bound.Set(changed)
		require.False(t, ok, "%s: value to value is not Set", bound.Keyword)
		_, ok = bound.Unset(changed)
		require.False(t, ok, "%s: value to value is not Unset", bound.Keyword)
	}
}
