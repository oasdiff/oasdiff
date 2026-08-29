package diff

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// Fields of openapi3.Schema that getSchemaDiff deliberately does not compare,
// with the reason.
var schemaFieldsNotDiffed = map[string]string{
	"Origin": "source-location bookkeeping, not document content",
}

// TestSchemaFieldCoverage sets each exported field of openapi3.Schema, one at
// a time, on one side of an otherwise-empty schema pair and requires the diff
// to see the change. A field kin adds is set like any other, so a field the
// diff does not compare fails the build here instead of shipping as a wrong
// verdict. Same gate as flatten/allof's
// TestMerge_SingleSchema_PreservesEveryField, asked of the diff.
func TestSchemaFieldCoverage(t *testing.T) {
	cfg := NewConfig()
	for f := range reflect.TypeFor[openapi3.Schema]().Fields() {
		if !f.IsExported() {
			continue
		}
		if _, waived := schemaFieldsNotDiffed[f.Name]; waived {
			continue
		}

		revision := &openapi3.Schema{}
		if !setNonZeroValue(reflect.ValueOf(revision).Elem().FieldByName(f.Name)) {
			// Silently skipping would reopen the hole this test exists to
			// close: a field of a kind setNonZeroValue cannot build would go
			// unchecked and the test would still pass.
			t.Errorf("cannot populate %s (%s); extend setNonZeroValue so the field is covered", f.Name, f.Type)
			continue
		}

		d, err := getSchemaDiff(cfg, newState(),
			&openapi3.SchemaRef{Value: &openapi3.Schema{}},
			&openapi3.SchemaRef{Value: revision})
		require.NoError(t, err)
		if d.Empty() {
			t.Errorf("openapi3.Schema.%s changed but the diff is empty\n  compare it in getSchemaDiff or add it to schemaFieldsNotDiffed with a reason", f.Name)
		}
	}
}

// setNonZeroValue fills v with a value the diff can tell from the zero
// schema's, derived from the field's type rather than listed per field.
func setNonZeroValue(v reflect.Value) bool {
	if !v.CanSet() {
		return false
	}
	// A schema position needs a usable schema, not just a non-nil pointer:
	// the diff dereferences it.
	if v.Type() == reflect.TypeFor[*openapi3.SchemaRef]() {
		v.Set(reflect.ValueOf(&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}))
		return true
	}
	// An empty Types would compare equal to an absent one.
	if v.Type() == reflect.TypeFor[*openapi3.Types]() {
		v.Set(reflect.ValueOf(&openapi3.Types{"string"}))
		return true
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString("sample")
		return true
	case reflect.Bool:
		v.SetBool(true)
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
		return true
	case reflect.Float64:
		v.SetFloat(1)
		return true
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		return setNonZeroValue(v.Index(0))
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		val := reflect.New(v.Type().Elem()).Elem()
		if !setNonZeroValue(k) || !setNonZeroValue(val) {
			return false
		}
		m.SetMapIndex(k, val)
		v.Set(m)
		return true
	case reflect.Pointer:
		// Allocated, not filled: Schema points at itself through SchemaRef,
		// so descending would not terminate, and a non-nil pointer to a zero
		// value already differs from the absent one.
		v.Set(reflect.New(v.Type().Elem()))
		return true
	case reflect.Interface:
		v.Set(reflect.ValueOf("sample"))
		return true
	case reflect.Struct:
		populated := false
		for _, fv := range v.Fields() {
			if setNonZeroValue(fv) {
				populated = true
			}
		}
		return populated
	}
	return false
}
