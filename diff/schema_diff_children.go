package diff

import "reflect"

// This file holds what the schema graph knows about the shape of a
// SchemaDiff: which of its fields hold child diffs, and the three ways the
// graph reaches them. The graph machinery goes through these, so a new field
// that holds schema diffs is a change here and nowhere else.

// childFields indexes the fields of SchemaDiff that hold schema diffs, found
// by type so that a new such field is unrolled without being listed here.
// TestUnroll_CutsEveryChildField fails for a field that holds schema diffs in
// a form the unroll does not follow.
var childFields = findChildFields()

func findChildFields() []int {
	var fields []int
	for field := range reflect.TypeFor[SchemaDiff]().Fields() {
		switch field.Type {
		case reflect.TypeFor[*SchemaDiff](), reflect.TypeFor[*SchemasDiff](), reflect.TypeFor[*SubschemasDiff]():
			fields = append(fields, field.Index[0])
		}
	}
	return fields
}

// children returns the nodes the diff links to through its child fields.
func (diff *SchemaDiff) children() []*SchemaDiff {
	var children []*SchemaDiff
	fields := reflect.ValueOf(diff).Elem()
	for _, index := range childFields {
		switch child := fields.Field(index).Interface().(type) {
		case *SchemaDiff:
			if child != nil {
				children = append(children, child)
			}
		case *SchemasDiff:
			if child != nil {
				for _, node := range child.Modified {
					children = append(children, node)
				}
			}
		case *SubschemasDiff:
			if child != nil {
				for _, modified := range child.Modified {
					children = append(children, modified.Diff)
				}
			}
		}
	}
	return children
}

// pendingSubschemas returns the diff's subschema lists whose matching has not
// run yet, the lists whose pairs the unroll compares.
func (diff *SchemaDiff) pendingSubschemas() []*SubschemasDiff {
	var pending []*SubschemasDiff
	fields := reflect.ValueOf(diff).Elem()
	for _, index := range childFields {
		if subschemas, ok := fields.Field(index).Interface().(*SubschemasDiff); ok && subschemas != nil && subschemas.pending {
			pending = append(pending, subschemas)
		}
	}
	return pending
}

// childVisitor is what a walk of the graph does to a node's children. It is
// an interface so that the shape of SchemaDiff does not depend on the walk;
// *unroller satisfies it with the methods it already has.
type childVisitor interface {
	unroll(*SchemaDiff) *SchemaDiff
	schemas(*SchemasDiff) *SchemasDiff
	subschemas(*SubschemasDiff) *SubschemasDiff
}

// withChildren returns a copy of the diff whose child fields hold what the
// visitor returned for each of them.
func (diff *SchemaDiff) withChildren(visitor childVisitor) SchemaDiff {
	result := *diff
	fields := reflect.ValueOf(&result).Elem()
	for _, index := range childFields {
		field := fields.Field(index)
		switch child := field.Interface().(type) {
		case *SchemaDiff:
			field.Set(reflect.ValueOf(visitor.unroll(child)))
		case *SchemasDiff:
			field.Set(reflect.ValueOf(visitor.schemas(child)))
		case *SubschemasDiff:
			field.Set(reflect.ValueOf(visitor.subschemas(child)))
		}
	}
	return result
}
