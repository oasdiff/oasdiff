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

					appendResultItem := func(messageId string, a ...any) {
						result = append(result, NewApiChange(
							messageId,
							config,
							a,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithDetails(mediaTypeDetails))
					}

					if mediaTypeDiff.SchemaDiff.PatternPropertiesDiff != nil {
						for _, pattern := range mediaTypeDiff.SchemaDiff.PatternPropertiesDiff.Added {
							appendResultItem(ResponseBodyPatternPropertyAddedId, pattern, responseStatus)
						}
						for _, pattern := range mediaTypeDiff.SchemaDiff.PatternPropertiesDiff.Deleted {
							appendResultItem(ResponseBodyPatternPropertyRemovedId, pattern, responseStatus)
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff.PatternPropertiesDiff == nil {
								return
							}
							propName := propertyFullName(propertyPath, propertyName)
							for _, pattern := range propertyDiff.PatternPropertiesDiff.Added {
								appendResultItem(ResponsePropertyPatternPropertyAddedId, pattern, propName, responseStatus)
							}
							for _, pattern := range propertyDiff.PatternPropertiesDiff.Deleted {
								appendResultItem(ResponsePropertyPatternPropertyRemovedId, pattern, propName, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
