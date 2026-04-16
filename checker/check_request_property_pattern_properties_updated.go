package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPatternPropertyAddedId       = "request-body-pattern-property-added"
	RequestBodyPatternPropertyRemovedId     = "request-body-pattern-property-removed"
	RequestPropertyPatternPropertyAddedId   = "request-property-pattern-property-added"
	RequestPropertyPatternPropertyRemovedId = "request-property-pattern-property-removed"
)

func RequestPropertyPatternPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				if mediaTypeDiff.SchemaDiff == nil {
					continue
				}

				if mediaTypeDiff.SchemaDiff.PatternPropertiesDiff != nil {
					patPropsDiff := mediaTypeDiff.SchemaDiff.PatternPropertiesDiff
					for _, pattern := range patPropsDiff.Added {
						revisionSource := SchemaMapItemSource(operationsSources, operationItem.Revision, patPropsDiff.Revision, pattern)
						result = append(result, NewApiChange(
							RequestBodyPatternPropertyAddedId,
							config,
							[]any{pattern},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					for _, pattern := range patPropsDiff.Deleted {
						baseSource := SchemaMapItemSource(operationsSources, operationItem.Base, patPropsDiff.Base, pattern)
						result = append(result, NewApiChange(
							RequestBodyPatternPropertyRemovedId,
							config,
							[]any{pattern},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.PatternPropertiesDiff == nil {
							return
						}
						propName := propertyFullName(propertyPath, propertyName)
						patPropsDiff := propertyDiff.PatternPropertiesDiff
						for _, pattern := range patPropsDiff.Added {
							revisionSource := SchemaMapItemSource(operationsSources, operationItem.Revision, patPropsDiff.Revision, pattern)
							result = append(result, NewApiChange(
								RequestPropertyPatternPropertyAddedId,
								config,
								[]any{pattern, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						for _, pattern := range patPropsDiff.Deleted {
							baseSource := SchemaMapItemSource(operationsSources, operationItem.Base, patPropsDiff.Base, pattern)
							result = append(result, NewApiChange(
								RequestPropertyPatternPropertyRemovedId,
								config,
								[]any{pattern, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
