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
	return NonZeroScale(v, tag, 1)
}

// NonZeroScale is NonZero with the numeric sample set to scale, so two calls
// with different scales produce ordered values for numeric kinds. Non-numeric
// kinds ignore the scale.
func NonZeroScale(v reflect.Value, tag string, scale uint64) bool {
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
		v.SetInt(int64(scale))
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(scale)
		return true
	case reflect.Float64:
		v.SetFloat(float64(scale))
		return true
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		return NonZeroScale(v.Index(0), tag, scale)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		val := reflect.New(v.Type().Elem()).Elem()
		if !NonZeroScale(k, tag, scale) || !NonZeroScale(val, tag, scale) {
			return false
		}
		m.SetMapIndex(k, val)
		v.Set(m)
		return true
	case reflect.Pointer:
		// Scalar pointees are filled; struct pointees are allocated but left
		// zero, because Schema points at itself through SchemaRef, so
		// descending would not terminate, and a non-nil pointer to a zero
		// value already differs from the absent one.
		p := reflect.New(v.Type().Elem())
		switch v.Type().Elem().Kind() {
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float64:
			NonZeroScale(p.Elem(), tag, scale)
		}
		v.Set(p)
		return true
	case reflect.Interface:
		v.Set(reflect.ValueOf("v-" + tag))
		return true
	case reflect.Struct:
		populated := false
		for _, fv := range v.Fields() {
			if NonZeroScale(fv, tag, scale) {
				populated = true
			}
		}
		return populated
	}
	return false
}
