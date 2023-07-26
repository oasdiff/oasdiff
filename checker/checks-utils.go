package checker

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
)

func propertyFullName(propertyPath string, propertyNames ...string) string {
	propertyFullName := strings.Join(propertyNames, "/")
	if propertyPath != "" {
		propertyFullName = propertyPath + "/" + propertyFullName
	}
	return propertyFullName
}

func colorizedValue(arg interface{}) string {
	str := interfaceToString(arg)
	if isPipedOutput() {
		return fmt.Sprintf("'%s'", str)
	}
	return color.InBold(fmt.Sprintf("'%s'", str))
}

func interfaceToString(arg interface{}) string {
	if arg == nil {
		return "undefined"
	}

	argString, ok := arg.(string)
	if ok {
		return argString
	}

	argUint64, ok := arg.(uint64)
	if ok {
		return fmt.Sprintf("%d", argUint64)
	}

	argFloat64, ok := arg.(float64)
	if ok {
		return fmt.Sprintf("%.2f", argFloat64)
	}

	return fmt.Sprintf("%s", arg)
}

func checkModifiedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *diff.SchemaDiff, propertyParentItem *diff.SchemaDiff)) {
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
		for k, v := range schemaDiff.AllOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for k, v := range schemaDiff.AnyOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for k, v := range schemaDiff.OneOfDiff.Modified {
			processModifiedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
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
}

func checkAddedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}
	processAddedPropertiesDiff("", "", schemaDiff, nil, processor)
}

func processAddedPropertiesDiff(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if propertyName != "" {
		if propertyPath == "" {
			propertyPath = propertyName
		} else {
			propertyPath = propertyPath + "/" + propertyName
		}
	}

	if schemaDiff.AllOfDiff != nil {
		for k, v := range schemaDiff.AllOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.AnyOfDiff != nil {
		for k, v := range schemaDiff.AnyOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for k, v := range schemaDiff.OneOfDiff.Modified {
			processAddedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processAddedPropertiesDiff(fmt.Sprintf("%s/items", propertyPath), "", schemaDiff.ItemsDiff, schemaDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Added {
			processor(propertyPath, v, schemaDiff.Revision.Value.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processAddedPropertiesDiff(propertyPath, i, v, schemaDiff, processor)
		}
	}
}

func checkDeletedPropertiesDiff(schemaDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if schemaDiff == nil {
		return
	}

	processDeletedPropertiesDiff("", "", schemaDiff, nil, processor)
}

func processDeletedPropertiesDiff(propertyPath string, propertyName string, schemaDiff *diff.SchemaDiff, parentDiff *diff.SchemaDiff, processor func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, propertyParentDiff *diff.SchemaDiff)) {
	if propertyName != "" {
		if propertyPath == "" {
			propertyPath = propertyName
		} else {
			propertyPath = propertyPath + "/" + propertyName
		}
	}

	if schemaDiff.AllOfDiff != nil {
		for k, v := range schemaDiff.AllOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/allOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}
	if schemaDiff.AnyOfDiff != nil {
		for k, v := range schemaDiff.AnyOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/anyOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.OneOfDiff != nil {
		for k, v := range schemaDiff.OneOfDiff.Modified {
			processDeletedPropertiesDiff(fmt.Sprintf("%s/oneOf[%s]", propertyPath, k), "", v, schemaDiff, processor)
		}
	}

	if schemaDiff.ItemsDiff != nil {
		processDeletedPropertiesDiff(fmt.Sprintf("%s/items", propertyPath), "", schemaDiff.ItemsDiff, schemaDiff, processor)
	}

	if schemaDiff.PropertiesDiff != nil {
		for _, v := range schemaDiff.PropertiesDiff.Deleted {
			processor(propertyPath, v, schemaDiff.Base.Value.Properties[v].Value, schemaDiff)
		}
		for i, v := range schemaDiff.PropertiesDiff.Modified {
			processDeletedPropertiesDiff(propertyPath, i, v, schemaDiff, processor)
		}
	}
}

func isIncreased(from interface{}, to interface{}) bool {
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

func isIncreasedValue(diff *diff.ValueDiff) bool {
	return isIncreased(diff.From, diff.To)
}

func isDecreasedValue(diff *diff.ValueDiff) bool {
	return isDecreased(diff.From, diff.To)
}

func isDecreased(from interface{}, to interface{}) bool {
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

func empty2none(a interface{}) interface{} {
	if a == nil || a == "" {
		return colorizedValue("none")
	}
	return colorizedValue(a)
}
