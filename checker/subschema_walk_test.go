package checker

import (
	"reflect"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// notTraversed are SchemaDiff fields that hold a sub-schema comparison the walk
// deliberately does not descend into, with the reason. A field that is neither
// walked nor listed here fails the test below.
var notTraversed = map[string]string{
	"DefsDiff": "$defs is a namespace of schemas reached through $ref, so a change to one is reported where it is used; walking it as well would report the change twice",
}

// A sub-schema field added to SchemaDiff has to be added to subschemaWalk, or
// every check built on the walk silently stops seeing changes under it. That is
// a quiet failure: the checks keep passing, on less of the document.
//
// Each field is filled on its own and the walk is run over it, so what fails is
// a sub-schema the walk never reaches.
func TestSubschemaTraversalIsComplete(t *testing.T) {
	for _, name := range subschemaFields() {
		t.Run(name, func(t *testing.T) {
			target := &diff.SchemaDiff{}
			root := &diff.SchemaDiff{}
			putSubschema(t, root, name, target)

			reached := false
			subschemaWalk{enter: func(_ string, _ string, schemaDiff *diff.SchemaDiff, _ *diff.SchemaDiff, _ bool) {
				if schemaDiff == target {
					reached = true
				}
			}}.walk("", "", root, nil, false)

			if reason, ok := notTraversed[name]; ok {
				require.Falsef(t, reached,
					"SchemaDiff.%s is listed in notTraversed but the walk descends into it\n  reason given: %s", name, reason)
				return
			}
			require.Truef(t, reached,
				"subschemaWalk does not descend into SchemaDiff.%s\n"+
					"  every check built on the walk will miss changes under it\n"+
					"  add it to the walk, or to notTraversed with a reason", name)
		})
	}
}

// putSubschema sets the named SchemaDiff field to a value holding target, in
// whichever shape that field uses to carry sub-schemas.
func putSubschema(t *testing.T, schemaDiff *diff.SchemaDiff, field string, target *diff.SchemaDiff) {
	t.Helper()

	f := reflect.ValueOf(schemaDiff).Elem().FieldByName(field)
	require.Truef(t, f.IsValid(), "SchemaDiff has no field %s", field)

	switch f.Type() {
	case reflect.TypeFor[*diff.SchemaDiff]():
		f.Set(reflect.ValueOf(target))
	case reflect.TypeFor[*diff.SubschemasDiff]():
		f.Set(reflect.ValueOf(&diff.SubschemasDiff{Modified: diff.ModifiedSubschemas{{Diff: target}}}))
	case reflect.TypeFor[*diff.SchemasDiff]():
		f.Set(reflect.ValueOf(&diff.SchemasDiff{Modified: diff.ModifiedSchemasMap{"subschema": target}}))
	default:
		require.Failf(t, "unknown sub-schema shape",
			"SchemaDiff.%s is a %s, which this test cannot fill\n"+
				"  teach putSubschema to build one, so the walk is checked against it too", field, f.Type())
	}
}

// subschemaFields returns the SchemaDiff fields that hold a comparison of one or
// more sub-schemas, which are the ones a traversal has to descend into.
func subschemaFields() []string {
	var names []string
	for f := range reflect.TypeFor[diff.SchemaDiff]().Fields() {
		if !f.IsExported() || !strings.HasSuffix(f.Name, "Diff") {
			continue
		}
		if holdsSubschemas(f.Type, map[reflect.Type]bool{}) {
			names = append(names, f.Name)
		}
	}
	return names
}

// holdsSubschemas reports whether a type is, or contains, a SchemaDiff. seen
// stops the recursion on the cycles SchemaDiff's own fields form.
func holdsSubschemas(t reflect.Type, seen map[reflect.Type]bool) bool {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	if t == reflect.TypeFor[diff.SchemaDiff]() {
		return true
	}
	if t.Kind() != reflect.Struct || seen[t] {
		return false
	}
	seen[t] = true
	for f := range t.Fields() {
		if holdsSubschemas(f.Type, seen) {
			return true
		}
	}
	return false
}
