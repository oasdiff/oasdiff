package checker

import (
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

					schemaDiff := paramDiff.SchemaDiff
					typeDiff := schemaDiff.TypeDiff
					formatDiff := schemaDiff.FormatDiff

					if !typeDiff.Empty() || !formatDiff.Empty() {

						// Suppress null-only type changes (handled by nullable checkers)
						if isNullTypeChange(typeDiff) && formatDiff.Empty() {
							continue
						}

						id := RequestParameterTypeGeneralizedId

						if breakingTypeFormatChangedInRequest(typeDiff, formatDiff, false, schemaDiff) {
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
						))
					}

					CheckModifiedPropertiesDiff(
						schemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

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
								))
							}
						})
				}
			}
		}
	}
	return result
}
