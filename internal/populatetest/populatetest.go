// Package populatetest fills reflect values with non-zero samples derived
// from their types. Only tests import it.
package populatetest

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

// NonZero fills v with a value distinguishable from the zero value, derived
// from v's type. String samples include the tag, so a value traces back to
// the field it was built for. It reports false for a kind it cannot build.
func NonZero(v reflect.Value, tag string) bool {
	if !v.CanSet() {
		return false
	}
	// A full schema rather than a bare non-nil pointer, so the value
	// survives being dereferenced and compared.
	if v.Type() == reflect.TypeFor[*openapi3.SchemaRef]() {
		v.Set(reflect.ValueOf(&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}}}))
		return true
	}
	// An empty Types is semantically "no type".
	if v.Type() == reflect.TypeFor[*openapi3.Types]() {
		v.Set(reflect.ValueOf(&openapi3.Types{"object"}))
		return true
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString("v-" + tag)
		return true
	case reflect.Bool:
		v.SetBool(true)
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
		return true
	case reflect.Float64:
		v.SetFloat(1)
		return true
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		return NonZero(v.Index(0), tag)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		val := reflect.New(v.Type().Elem()).Elem()
		if !NonZero(k, tag) || !NonZero(val, tag) {
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
		v.Set(reflect.ValueOf("v-" + tag))
		return true
	case reflect.Struct:
		populated := false
		for _, fv := range v.Fields() {
			if NonZero(fv, tag) {
				populated = true
			}
		}
		return populated
	}
	return false
}
