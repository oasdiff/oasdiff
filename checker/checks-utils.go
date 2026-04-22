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

// joinPath joins a property path with a segment using "/" as separator,
// avoiding a leading slash when the path is empty.
func joinPath(propertyPath, segment string) string {
	if propertyPath == "" {
		return segment
	}
	return propertyPath + "/" + segment
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
			processModifiedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processModifiedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processModifiedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, schemaDiff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processModifiedPropertiesDiff(joinPath(propertyPath, "items"), "", schemaDiff.ItemsDiff, schemaDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processModifiedPropertiesDiff(propertyPath, i, v, schemaDiff, processor)
		}
	}

	if schemaDiff.AdditionalPropertiesDiff != nil {
		processModifiedPropertiesDiff(joinPath(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, schemaDiff, processor)
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
			processAddedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processAddedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processAddedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processAddedPropertiesDiff(joinPath(propertyPath, "items"), "", schemaDiff.ItemsDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Added {
			processor(propertyPath, v, schemaDiff.Revision.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processAddedPropertiesDiff(propertyPath, i, v, processor)
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
			processDeletedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("allOf[%s]", v)), "", v.Diff, processor)
		}
	}
	if schemaDiff.AnyOfDiff != nil {
		for _, v := range schemaDiff.AnyOfDiff.Modified {
			processDeletedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("anyOf[%s]", v)), "", v.Diff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for _, v := range schemaDiff.OneOfDiff.Modified {
			processDeletedPropertiesDiff(joinPath(propertyPath, fmt.Sprintf("oneOf[%s]", v)), "", v.Diff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processDeletedPropertiesDiff(joinPath(propertyPath, "items"), "", schemaDiff.ItemsDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Deleted {
			processor(propertyPath, v, schemaDiff.Base.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processDeletedPropertiesDiff(propertyPath, i, v, processor)
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
