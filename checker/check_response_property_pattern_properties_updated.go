package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPatternPropertyAddedId       = "response-body-pattern-property-added"
	ResponseBodyPatternPropertyRemovedId     = "response-body-pattern-property-removed"
	ResponsePropertyPatternPropertyAddedId   = "response-property-pattern-property-added"
	ResponsePropertyPatternPropertyRemovedId = "response-property-pattern-property-removed"
)

func ResponsePropertyPatternPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.ContentDiff == nil || responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
					if mediaTypeDiff.SchemaDiff == nil {
						continue
					}

					if mediaTypeDiff.SchemaDiff.PatternPropertiesDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "patternProperties")
						for _, pattern := range mediaTypeDiff.SchemaDiff.PatternPropertiesDiff.Added {
							result = append(result, NewApiChange(
								ResponseBodyPatternPropertyAddedId,
								config,
								[]any{pattern, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						for _, pattern := range mediaTypeDiff.SchemaDiff.PatternPropertiesDiff.Deleted {
							result = append(result, NewApiChange(
								ResponseBodyPatternPropertyRemovedId,
								config,
								[]any{pattern, responseStatus},
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
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "patternProperties")
							for _, pattern := range propertyDiff.PatternPropertiesDiff.Added {
								result = append(result, NewApiChange(
									ResponsePropertyPatternPropertyAddedId,
									config,
									[]any{pattern, propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							for _, pattern := range propertyDiff.PatternPropertiesDiff.Deleted {
								result = append(result, NewApiChange(
									ResponsePropertyPatternPropertyRemovedId,
									config,
									[]any{pattern, propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
							}
						})
				}
			}
		}
	}
	return result
}
