package diff

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// reachesSchemaDiff reports whether a value of type t can hold a *SchemaDiff
// somewhere below it.
func reachesSchemaDiff(t reflect.Type, seen map[reflect.Type]bool) bool {
	if t == reflect.TypeFor[*SchemaDiff]() {
		return true
	}
	if seen[t] {
		return false
	}
	seen[t] = true
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		return reachesSchemaDiff(t.Elem(), seen)
	case reflect.Struct:
		for field := range t.Fields() {
			if reachesSchemaDiff(field.Type, seen) {
				return true
			}
		}
	}
	return false
}

// Every field of SchemaDiff that can hold a schema diff is a place the graph
// can cycle through, so the unroll must follow it and cut it. A node linked
// to itself through such a field must unroll with that field empty and the
// node's own change kept.
func TestUnroll_CutsEveryChildField(t *testing.T) {
	schemaDiffType := reflect.TypeFor[SchemaDiff]()
	for i := range schemaDiffType.NumField() {
		field := schemaDiffType.Field(i)
		if field.Name == "Base" || field.Name == "Revision" || !reachesSchemaDiff(field.Type, map[reflect.Type]bool{}) {
			continue
		}
		t.Run(field.Name, func(t *testing.T) {
			node := &SchemaDiff{DescriptionDiff: &ValueDiff{From: "a", To: "b"}}
			var link reflect.Value
			switch field.Type {
			case reflect.TypeFor[*SchemaDiff]():
				link = reflect.ValueOf(node)
			case reflect.TypeFor[*SchemasDiff]():
				link = reflect.ValueOf(&SchemasDiff{Modified: ModifiedSchemasMap{"self": node}})
			case reflect.TypeFor[*SubschemasDiff]():
				link = reflect.ValueOf(&SubschemasDiff{Modified: ModifiedSubschemas{{Diff: node}}})
			default:
				t.Fatalf("field %s has type %s, which can hold a schema diff: extend unroller.copy and this test", field.Name, field.Type)
			}
			reflect.ValueOf(node).Elem().Field(i).Set(link)

			graph := newSchemaGraph()
			unrolled := graph.unroll(node)
			require.NotNil(t, unrolled, "the node's own change must be kept")
			require.True(t, reflect.ValueOf(unrolled).Elem().Field(i).IsNil(), "the cycle through %s must be cut", field.Name)
			require.Equal(t, "b", unrolled.DescriptionDiff.To)
		})
	}
}

// A node reached on two paths unrolls to one shared diff when nothing below
// it cuts into a node above it.
func TestUnroll_SharesCutFreeNodes(t *testing.T) {
	leaf := &SchemaDiff{DescriptionDiff: &ValueDiff{From: "a", To: "b"}}
	root := &SchemaDiff{PropertiesDiff: &SchemasDiff{Modified: ModifiedSchemasMap{"x": leaf, "y": leaf}}}

	graph := newSchemaGraph()
	unrolled := graph.unroll(root)
	require.Same(t, unrolled.PropertiesDiff.Modified["x"], unrolled.PropertiesDiff.Modified["y"])
}
