package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterTypeChangedId                = "request-parameter-type-changed"
	RequestParameterTypeGeneralizedId            = "request-parameter-type-generalized"
	RequestParameterPropertyTypeChangedId        = "request-parameter-property-type-changed"
	RequestParameterPropertyTypeGeneralizedId    = "request-parameter-property-type-generalized"
	RequestParameterPropertyTypeSpecializedId    = "request-parameter-property-type-specialized"
	RequestParameterPropertyTypeChangedCommentId = "request-parameter-property-type-changed-warn-comment"
)

// isParameterScalarToFormExplodeArray reports whether typeDiff describes a
// scalar-to-array change on a parameter using form/explode serialization
// (the default for query and cookie parameters), where the new array's
// element type matches the old scalar type. Per the OpenAPI spec, a client
// sending a single value (?color=red) is a valid one-element array under
// these serialization rules, so the schema change is backwards-compatible.
func isParameterScalarToFormExplodeArray(paramDiff *diff.ParameterDiff, typeDiff *diff.StringsDiff) bool {
	if paramDiff == nil || paramDiff.Revision == nil || paramDiff.SchemaDiff == nil {
		return false
	}
	if typeDiff == nil ||
		len(typeDiff.Added) != 1 || typeDiff.Added[0] != "array" ||
		len(typeDiff.Deleted) != 1 || typeDiff.Deleted[0] == "array" {
		return false
	}

	method, err := paramDiff.Revision.SerializationMethod()
	if err != nil || method == nil ||
		method.Style != openapi3.SerializationForm || !method.Explode {
		return false
	}

	revSchema := paramDiff.SchemaDiff.Revision
	if revSchema == nil || revSchema.Items == nil || revSchema.Items.Value == nil {
		return false
	}
	itemTypes := revSchema.Items.Value.Type.Slice()
	return len(itemTypes) == 1 && itemTypes[0] == typeDiff.Deleted[0]
}

func RequestParameterTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil {
				continue
			}

			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {
				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}

					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "type")
					schemaDiff := paramDiff.SchemaDiff
					typeDiff := schemaDiff.TypeDiff
					formatDiff := schemaDiff.FormatDiff

					if !typeDiff.Empty() || !formatDiff.Empty() {

						// Suppress null-only type changes (handled by nullable checkers)
						if isNullTypeChange(typeDiff) && formatDiff.Empty() {
							continue
						}

						id := RequestParameterTypeGeneralizedId

						if typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff) &&
							!isParameterScalarToFormExplodeArray(paramDiff, typeDiff) {
							id = RequestParameterTypeChangedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff)},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}

					CheckModifiedPropertiesDiff(
						schemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "type")
							schemaDiff := propertyDiff
							typeDiff := schemaDiff.TypeDiff
							formatDiff := schemaDiff.FormatDiff

							if !typeDiff.Empty() || !formatDiff.Empty() {

								// Suppress null-only type changes (handled by nullable checkers)
								if isNullTypeChange(typeDiff) && formatDiff.Empty() {
									return
								}

								id, comment := checkRequestParameterPropertyTypeChanged(typeDiff, formatDiff, schemaDiff)

								result = append(result, NewApiChange(
									id,
									config,
									[]any{paramLocation, paramName, propertyFullName(propertyPath, propertyName), getBaseType(schemaDiff), getBaseFormat(schemaDiff), getRevisionType(schemaDiff), getRevisionFormat(schemaDiff)},
									comment,
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource))
							}
						})
				}
			}
		}
	}
	return result
}
