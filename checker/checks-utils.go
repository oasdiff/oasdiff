package checker

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

func commentId(id string) string {
	return id + "-comment"
}

func descriptionId(id string) string {
	return id + "-description"
}

func propertyFullName(propertyPath string, propertyNames ...string) string {
	propertyFullName := strings.Join(propertyNames, "/")
	if propertyPath != "" {
		propertyFullName = propertyPath + "/" + propertyFullName
	}
	return propertyFullName
}

func interfaceToString(arg any) string {
	if arg == nil {
		return "undefined"
	}

	if argString, ok := arg.(string); ok {
		return argString
	}

	if argStringList, ok := arg.([]string); ok {
		return strings.Join(argStringList, ", ")
	}

	if argInt, ok := arg.(int); ok {
		return fmt.Sprintf("%d", argInt)
	}

	if argUint64, ok := arg.(uint64); ok {
		return fmt.Sprintf("%d", argUint64)
	}

	if argFloat64, ok := arg.(float64); ok {
		return fmt.Sprintf("%.2f", argFloat64)
	}

	if argBool, ok := arg.(bool); ok {
		return fmt.Sprintf("%t", argBool)
	}

	return fmt.Sprintf("%v", arg)
}

func CheckModifiedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	processModifiedPropertiesDiff("", "", schemaDiff, nil, processor)
}

func processModifiedPropertiesDiff(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
	if propertyName != "" || propertyPath != "" {
		processor(propertyPath, propertyName, schemaDiff, parentDiff)
	}

	if propertyName != "" {
		if propertyPath == "" {
			propertyPath = propertyName
		} else {
			propertyPath = propertyPath + "/" + propertyName
		}
	}

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, v), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, v), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, v), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/items", propertyPath), "", schemaDiff.ItemsDiff, schemaDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processModifiedPropertiesDiff(propertyPath, i, v, schemaDiff, processor)
		}
	}

	if schemaDiff.AdditionalPropertiesDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/additionalProperties", propertyPath), "", schemaDiff.AdditionalPropertiesDiff, schemaDiff, processor)
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schema fields
	if schemaDiff.PrefixItemsDiff != nil {
		for _, v := range schemaDiff.PrefixItemsDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/prefixItems[%s]", propertyPath, v), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.ContainsDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/contains", propertyPath), "", schemaDiff.ContainsDiff, schemaDiff, processor)
	}

	if schemaDiff.PropertyNamesDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/propertyNames", propertyPath), "", schemaDiff.PropertyNamesDiff, schemaDiff, processor)
	}

	if schemaDiff.UnevaluatedItemsDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/unevaluatedItems", propertyPath), "", schemaDiff.UnevaluatedItemsDiff, schemaDiff, processor)
	}

	if schemaDiff.UnevaluatedPropertiesDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/unevaluatedProperties", propertyPath), "", schemaDiff.UnevaluatedPropertiesDiff, schemaDiff, processor)
	}

	if schemaDiff.IfDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/if", propertyPath), "", schemaDiff.IfDiff, schemaDiff, processor)
	}

	if schemaDiff.ThenDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/then", propertyPath), "", schemaDiff.ThenDiff, schemaDiff, processor)
	}

	if schemaDiff.ElseDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/else", propertyPath), "", schemaDiff.ElseDiff, schemaDiff, processor)
	}

	if schemaDiff.ContentSchemaDiff != nil {
		processModifiedPropertiesDiff(fmt.Sprintf("%s/contentSchema", propertyPath), "", schemaDiff.ContentSchemaDiff, schemaDiff, processor)
	}

	if schemaDiff.PatternPropertiesDiff != nil {
		for i, v := range schemaDiff.PatternPropertiesDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for i, v := range schemaDiff.DependentSchemasDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, schemaDiff, processor)
		}
	}
}

func CheckAddedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}
	processAddedPropertiesDiff("", "", schemaDiff, processor)
}

func processAddedPropertiesDiff(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if propertyName != "" {
		if propertyPath == "" {
			propertyPath = propertyName
		} else {
			propertyPath = propertyPath + "/" + propertyName
		}
	}

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/items", propertyPath), "", schemaDiff.ItemsDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Added {
			processor(propertyPath, v, schemaDiff.Revision.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processAddedPropertiesDiff(propertyPath, i, v, processor)
		}
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schema fields
	if schemaDiff.PrefixItemsDiff != nil {
		for _, v := range schemaDiff.PrefixItemsDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/prefixItems[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.ContainsDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/contains", propertyPath), "", schemaDiff.ContainsDiff, processor)
	}

	if schemaDiff.PropertyNamesDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/propertyNames", propertyPath), "", schemaDiff.PropertyNamesDiff, processor)
	}

	if schemaDiff.UnevaluatedItemsDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/unevaluatedItems", propertyPath), "", schemaDiff.UnevaluatedItemsDiff, processor)
	}

	if schemaDiff.UnevaluatedPropertiesDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/unevaluatedProperties", propertyPath), "", schemaDiff.UnevaluatedPropertiesDiff, processor)
	}

	if schemaDiff.IfDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/if", propertyPath), "", schemaDiff.IfDiff, processor)
	}

	if schemaDiff.ThenDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/then", propertyPath), "", schemaDiff.ThenDiff, processor)
	}

	if schemaDiff.ElseDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/else", propertyPath), "", schemaDiff.ElseDiff, processor)
	}

	if schemaDiff.ContentSchemaDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/contentSchema", propertyPath), "", schemaDiff.ContentSchemaDiff, processor)
	}

	if schemaDiff.PatternPropertiesDiff != nil {
		for i, v := range schemaDiff.PatternPropertiesDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, processor)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for i, v := range schemaDiff.DependentSchemasDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, processor)
		}
	}
}

func CheckDeletedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	processDeletedPropertiesDiff("", "", schemaDiff, processor)
}

func processDeletedPropertiesDiff(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if propertyName != "" {
		if propertyPath == "" {
			propertyPath = propertyName
		} else {
			propertyPath = propertyPath + "/" + propertyName
		}
	}

	if schemaDiff.AllOfDiff != nil {
		for _, v := range schemaDiff.AllOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}
	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/items", propertyPath), "", schemaDiff.ItemsDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Deleted {
			processor(propertyPath, v, schemaDiff.Base.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processDeletedPropertiesDiff(propertyPath, i, v, processor)
		}
	}

	// OpenAPI 3.1 / JSON Schema 2020-12 sub-schema fields
	if schemaDiff.PrefixItemsDiff != nil {
		for _, v := range schemaDiff.PrefixItemsDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/prefixItems[%s]", propertyPath, v), "", v.Diff, processor)
		}
	}

	if schemaDiff.ContainsDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/contains", propertyPath), "", schemaDiff.ContainsDiff, processor)
	}

	if schemaDiff.PropertyNamesDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/propertyNames", propertyPath), "", schemaDiff.PropertyNamesDiff, processor)
	}

	if schemaDiff.UnevaluatedItemsDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/unevaluatedItems", propertyPath), "", schemaDiff.UnevaluatedItemsDiff, processor)
	}

	if schemaDiff.UnevaluatedPropertiesDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/unevaluatedProperties", propertyPath), "", schemaDiff.UnevaluatedPropertiesDiff, processor)
	}

	if schemaDiff.IfDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/if", propertyPath), "", schemaDiff.IfDiff, processor)
	}

	if schemaDiff.ThenDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/then", propertyPath), "", schemaDiff.ThenDiff, processor)
	}

	if schemaDiff.ElseDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/else", propertyPath), "", schemaDiff.ElseDiff, processor)
	}

	if schemaDiff.ContentSchemaDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/contentSchema", propertyPath), "", schemaDiff.ContentSchemaDiff, processor)
	}

	if schemaDiff.PatternPropertiesDiff != nil {
		for i, v := range schemaDiff.PatternPropertiesDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/patternProperties[%s]", propertyPath, i), "", v, processor)
		}
	}

	if schemaDiff.DependentSchemasDiff != nil {
		for i, v := range schemaDiff.DependentSchemasDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/dependentSchemas[%s]", propertyPath, i), "", v, processor)
		}
	}
}

func IsIncreased(from any, to any) bool {
	fromUint64, ok := from.(uint64)
	toUint64, okTo := to.(uint64)
	if ok && okTo {
		return fromUint64 < toUint64
	}
	fromFloat64, ok := from.(float64)
	toFloat64, okTo := to.(float64)
	if ok && okTo {
		return fromFloat64 < toFloat64
	}
	return false
}

func IsIncreasedValue(diff *diff.ValueDiff) bool {
	return IsIncreased(diff.From, diff.To)
}

func IsDecreasedValue(diff *diff.ValueDiff) bool {
	return IsDecreased(diff.From, diff.To)
}

func IsDecreased(from any, to any) bool {
	fromUint64, ok := from.(uint64)
	toUint64, okTo := to.(uint64)
	if ok && okTo {
		return fromUint64 > toUint64
	}
	fromFloat64, ok := from.(float64)
	toFloat64, okTo := to.(float64)
	if ok && okTo {
		return fromFloat64 > toFloat64
	}
	return false
}
