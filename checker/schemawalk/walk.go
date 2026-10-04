package schemawalk

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// Walker recurses over a schema diff's sub-schemas once, so that a walk
// using it supplies only the part that differs: what it emits, and where.
type Walker struct {
	// Enter is called for the node itself, before its name is appended to the
	// path, so a caller receives the two separately.
	Enter func(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool)
	// Properties is called for a node's own properties, after its name is
	// appended, so the path already names the node they belong to.
	Properties func(propertyPath string, schemaDiff *diff.SchemaDiff, underAllOf bool)
	// A shared $ref can reach the same diff through many property paths.
	// Keep one representative path per parent and allOf context.
	seen map[walkVisit]struct{}
}

type walkVisit struct {
	schemaDiff *diff.SchemaDiff
	parentDiff *diff.SchemaDiff
	underAllOf bool
}

// Walk visits the sub-schemas of schemaDiff, starting from the schema itself.
func (w Walker) Walk(schemaDiff *diff.SchemaDiff) {
	w.walk("", "", schemaDiff, nil, false)
}

func (w Walker) walk(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool) {
	if w.seen == nil {
		w.seen = make(map[walkVisit]struct{})
	}
	visit := walkVisit{schemaDiff: schemaDiff, parentDiff: parentDiff, underAllOf: underAllOf}
	if _, ok := w.seen[visit]; ok {
		return
	}
	w.seen[visit] = struct{}{}

	if w.Enter != nil && (propertyName != "" || propertyPath != "") {
		w.Enter(propertyPath, propertyName, schemaDiff, parentDiff, underAllOf)
	}

	if propertyName != "" {
		propertyPath = PropertyFullName(propertyPath, propertyName)
	}

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			w.walk(PropertyFullName(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, schemaDiff, true)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			w.walk(PropertyFullName(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			w.walk(PropertyFullName(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		w.walk(PropertyFullName(propertyPath, "items"), "", schemaDiff.ItemsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PropertiesDiff != nil {
		if w.Properties != nil {
			w.Properties(propertyPath, schemaDiff, underAllOf)
		}
		for _, name := range slices.Sorted(maps.Keys(schemaDiff.PropertiesDiff.Modified)) {
			v := schemaDiff.PropertiesDiff.Modified[name]
			w.walk(propertyPath, name, v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.AdditionalPropertiesDiff != nil {
		w.walk(PropertyFullName(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, schemaDiff, underAllOf)
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schema fields
	if schemaDiff.PrefixItemsDiff != nil {
		for _, v := range schemaDiff.PrefixItemsDiff.Modified {
			w.walk(fmt.Sprintf("%s/prefixItems[%s]", propertyPath, v), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.ContainsDiff != nil {
		w.walk(fmt.Sprintf("%s/contains", propertyPath), "", schemaDiff.ContainsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PropertyNamesDiff != nil {
		w.walk(fmt.Sprintf("%s/propertyNames", propertyPath), "", schemaDiff.PropertyNamesDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.UnevaluatedItemsDiff != nil {
		w.walk(fmt.Sprintf("%s/unevaluatedItems", propertyPath), "", schemaDiff.UnevaluatedItemsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.UnevaluatedPropertiesDiff != nil {
		w.walk(fmt.Sprintf("%s/unevaluatedProperties", propertyPath), "", schemaDiff.UnevaluatedPropertiesDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.IfDiff != nil {
		w.walk(fmt.Sprintf("%s/if", propertyPath), "", schemaDiff.IfDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ThenDiff != nil {
		w.walk(fmt.Sprintf("%s/then", propertyPath), "", schemaDiff.ThenDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ElseDiff != nil {
		w.walk(fmt.Sprintf("%s/else", propertyPath), "", schemaDiff.ElseDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.NotDiff != nil {
		w.walk(fmt.Sprintf("%s/not", propertyPath), "", schemaDiff.NotDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.ContentSchemaDiff != nil {
		w.walk(fmt.Sprintf("%s/contentSchema", propertyPath), "", schemaDiff.ContentSchemaDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PatternPropertiesDiff != nil {
		for _, i := range slices.Sorted(maps.Keys(schemaDiff.PatternPropertiesDiff.Modified)) {
			v := schemaDiff.PatternPropertiesDiff.Modified[i]
			w.walk(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for _, i := range slices.Sorted(maps.Keys(schemaDiff.DependentSchemasDiff.Modified)) {
			v := schemaDiff.DependentSchemasDiff.Modified[i]
			w.walk(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}
}

func ModifiedProperties(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	Walker{Enter: func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff, _ bool) {
		processor(propertyPath, propertyName, propertyItem, propertyParentItem)
	}}.Walk(schemaDiff)
}

func AddedProperties(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff, underAllOf bool)) {
	if schemaDiff == nil {
		return
	}

	Walker{Properties: func(propertyPath string, sd *diff.SchemaDiff, underAllOf bool) {
		for _, name := range sd.PropertiesDiff.Added {
			processor(propertyPath, name, sd.Revision.Properties[name].Value, sd, underAllOf)
		}
	}}.Walk(schemaDiff)
}

func DeletedProperties(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff, underAllOf bool)) {
	if schemaDiff == nil {
		return
	}

	Walker{Properties: func(propertyPath string, sd *diff.SchemaDiff, underAllOf bool) {
		for _, name := range sd.PropertiesDiff.Deleted {
			processor(propertyPath, name, sd.Base.Properties[name].Value, sd, underAllOf)
		}
	}}.Walk(schemaDiff)
}

// PropertyFullName appends property names to a path, separated by "/". It is
// how the walk names a sub-schema, so a check that names one in a message
// uses it too.
func PropertyFullName(propertyPath string, propertyNames ...string) string {
	fullName := strings.Join(propertyNames, "/")
	if propertyPath != "" {
		fullName = propertyPath + "/" + fullName
	}
	return fullName
}
