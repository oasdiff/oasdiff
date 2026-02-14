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
						appendResultItem(RequestBodyPatternPropertyAddedId, pattern)
					}
					for _, pattern := range mediaTypeDiff.SchemaDiff.PatternPropertiesDiff.Deleted {
						appendResultItem(RequestBodyPatternPropertyRemovedId, pattern)
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
							appendResultItem(RequestPropertyPatternPropertyAddedId, pattern, propName)
						}
						for _, pattern := range propertyDiff.PatternPropertiesDiff.Deleted {
							appendResultItem(RequestPropertyPatternPropertyRemovedId, pattern, propName)
						}
					})
			}
		}
	}
	return result
}
