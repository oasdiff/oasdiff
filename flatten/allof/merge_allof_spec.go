package allof

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSpec merges all instances of allOf in place, across every schema in
// the document. Merge handles each schema's whole subtree (including its own
// cycle tracking), so the walk hands it each attachment point once and skips
// descent.
func MergeSpec(spec *openapi3.T) (*openapi3.T, error) {
	err := spec.WalkSchemas(func(_ string, s *openapi3.SchemaRef) error {
		m, err := Merge(*s)
		if err != nil {
			return err
		}
		// Every $ref to this schema shares one Value, so writing the merge
		// into it updates every use. Assigning s.Value would update only
		// this reference and leave the rest unmerged.
		*s.Value = *m
		// The copy leaves the subtree's self-references aimed at the object
		// Merge returned rather than the one just written into: a recursive
		// schema's one-object cycle becomes a two-object cycle, and later
		// attachment points then see two distinct originals for one schema,
		// which the merge cache cannot unify.
		redirectSchemaRefs(reflect.ValueOf(s.Value), m, s.Value, map[*openapi3.Schema]bool{})
		return openapi3.SkipSubtree
	})
	return spec, err
}

// redirectSchemaRefs walks every SchemaRef reachable from v and points those
// whose Value is from at to instead. The traversal is type-driven so a new
// schema field carrying subschemas is covered without being listed here.
func redirectSchemaRefs(v reflect.Value, from, to *openapi3.Schema, seen map[*openapi3.Schema]bool) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		if ref, ok := v.Interface().(*openapi3.SchemaRef); ok {
			if ref.Value == from {
				ref.Value = to
			}
			redirectSchemaRefs(reflect.ValueOf(ref.Value), from, to, seen)
			return
		}
		if schema, ok := v.Interface().(*openapi3.Schema); ok {
			if schema == nil || seen[schema] {
				return
			}
			seen[schema] = true
		}
		redirectSchemaRefs(v.Elem(), from, to, seen)
	case reflect.Slice:
		for i := range v.Len() {
			redirectSchemaRefs(v.Index(i), from, to, seen)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			redirectSchemaRefs(v.MapIndex(key), from, to, seen)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				redirectSchemaRefs(v.Field(i), from, to, seen)
			}
		}
	}
}
