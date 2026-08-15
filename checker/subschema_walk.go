package checker

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// subschemaWalk recurses over a schema diff's sub-schemas once, so that a walk
// using it supplies only the part that differs: what it emits, and where.
type subschemaWalk struct {
	// enter is called for the node itself, before its name is appended to the
	// path, so a caller receives the two separately.
	enter func(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool)
	// properties is called for a node's own properties, after its name is
	// appended, so the path already names the node they belong to.
	properties func(propertyPath string, schemaDiff *diff.SchemaDiff)
}

func (w subschemaWalk) walk(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, underAllOf bool) {
	if w.enter != nil && (propertyName != "" || propertyPath != "") {
		w.enter(propertyPath, propertyName, schemaDiff, parentDiff, underAllOf)
	}

	if propertyName != "" {
		propertyPath = propertyFullName(propertyPath, propertyName)
	}

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, schemaDiff, true)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			w.walk(propertyFullName(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		w.walk(propertyFullName(propertyPath, "items"), "", schemaDiff.ItemsDiff, schemaDiff, underAllOf)
	}

	if schemaDiff.PropertiesDiff != nil {
		if w.properties != nil {
			w.properties(propertyPath, schemaDiff)
		}
		for name, v := range schemaDiff.PropertiesDiff.Modified {
			w.walk(propertyPath, name, v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.AdditionalPropertiesDiff != nil {
		w.walk(propertyFullName(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, schemaDiff, underAllOf)
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
		for i, v := range schemaDiff.PatternPropertiesDiff.Modified {
			w.walk(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for i, v := range schemaDiff.DependentSchemasDiff.Modified {
			w.walk(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, schemaDiff, underAllOf)
		}
	}
}

func checkModifiedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{enter: func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff, _ bool) {
		processor(propertyPath, propertyName, propertyItem, propertyParentItem)
	}}.walk("", "", schemaDiff, nil, false)
}

func checkAddedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{properties: func(propertyPath string, sd *diff.SchemaDiff) {
		for _, name := range sd.PropertiesDiff.Added {
			processor(propertyPath, name, sd.Revision.Properties[name].Value, sd)
		}
	}}.walk("", "", schemaDiff, nil, false)
}

func checkDeletedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	subschemaWalk{properties: func(propertyPath string, sd *diff.SchemaDiff) {
		for _, name := range sd.PropertiesDiff.Deleted {
			processor(propertyPath, name, sd.Base.Properties[name].Value, sd)
		}
	}}.walk("", "", schemaDiff, nil, false)
}
