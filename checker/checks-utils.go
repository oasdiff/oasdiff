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

	if schemaDiff.AdditionalPropertiesDiff != nil {
		processAddedPropertiesDiff(joinPath(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, processor)
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

	if schemaDiff.AdditionalPropertiesDiff != nil {
		processDeletedPropertiesDiff(joinPath(propertyPath, "additionalProperties"), "", schemaDiff.AdditionalPropertiesDiff, processor)
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

// splitSubschemasByAnnotationOnly partitions subschemas into two disjoint
// sets according to whether the body at Subschema.Index in schemaRefs
// carries a validation-significant keyword.
//
//   - kept: subschemas with at least one constraint, type, property,
//     enum, etc. These drive the original allOf-changed emission at its
//     conventional severity (ERR / WARN for request, INFO for response).
//   - annotationOnly: subschemas that hold only annotation keywords
//     (title, description, examples, default, externalDocs, $comment).
//     Validation-equivalent to {} — they don't reject any previously-
//     valid instance. Callers emit them at INFO so the document-level
//     change stays auditable in the changelog without contaminating
//     breaking.
//
// Motivating case: handrews on OAS discussion #3793 — adding an
// `allOf: [{title: "..."}]` is not a breaking change the way adding a
// real constraint is, but the original allOf-added check fired ERR on it.
func splitSubschemasByAnnotationOnly(subschemas diff.Subschemas, schemaRefs openapi3.SchemaRefs) (kept diff.Subschemas, annotationOnly diff.Subschemas) {
	if len(subschemas) == 0 {
		return subschemas, nil
	}
	kept = make(diff.Subschemas, 0, len(subschemas))
	annotationOnly = make(diff.Subschemas, 0, len(subschemas))
	emptyRef := &openapi3.SchemaRef{Value: &openapi3.Schema{}}
	for _, s := range subschemas {
		isAnnotationOnly := false
		if s.Index >= 0 && s.Index < len(schemaRefs) {
			ref := schemaRefs[s.Index]
			if ref != nil && ref.Value != nil &&
				diff.SchemaRefsValidationEquivalent(diff.NewConfig(), emptyRef, ref) {
				isAnnotationOnly = true
			}
		}
		if isAnnotationOnly {
			annotationOnly = append(annotationOnly, s)
		} else {
			kept = append(kept, s)
		}
	}
	return kept, annotationOnly
}
