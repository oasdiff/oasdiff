package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyIfAddedId       = "response-body-if-added"
	ResponseBodyIfRemovedId     = "response-body-if-removed"
	ResponseBodyThenAddedId     = "response-body-then-added"
	ResponseBodyThenRemovedId   = "response-body-then-removed"
	ResponseBodyElseAddedId     = "response-body-else-added"
	ResponseBodyElseRemovedId   = "response-body-else-removed"
	ResponsePropertyIfAddedId   = "response-property-if-added"
	ResponsePropertyIfRemovedId = "response-property-if-removed"

	ResponsePropertyThenAddedId   = "response-property-then-added"
	ResponsePropertyThenRemovedId = "response-property-then-removed"
	ResponsePropertyElseAddedId   = "response-property-else-added"
	ResponsePropertyElseRemovedId = "response-property-else-removed"
)

func ResponsePropertyIfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					checkConditionalSchemaDiffResponse(mediaTypeDiff.SchemaDiff.IfDiff, ResponseBodyIfAddedId, ResponseBodyIfRemovedId, responseStatus, appendResultItem)
					checkConditionalSchemaDiffResponse(mediaTypeDiff.SchemaDiff.ThenDiff, ResponseBodyThenAddedId, ResponseBodyThenRemovedId, responseStatus, appendResultItem)
					checkConditionalSchemaDiffResponse(mediaTypeDiff.SchemaDiff.ElseDiff, ResponseBodyElseAddedId, ResponseBodyElseRemovedId, responseStatus, appendResultItem)

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							propName := propertyFullName(propertyPath, propertyName)

							appendPropResultItem := func(messageId string, a ...any) {
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

							checkConditionalSchemaDiffPropertyResponse(propertyDiff.IfDiff, ResponsePropertyIfAddedId, ResponsePropertyIfRemovedId, propName, responseStatus, appendPropResultItem)
							checkConditionalSchemaDiffPropertyResponse(propertyDiff.ThenDiff, ResponsePropertyThenAddedId, ResponsePropertyThenRemovedId, propName, responseStatus, appendPropResultItem)
							checkConditionalSchemaDiffPropertyResponse(propertyDiff.ElseDiff, ResponsePropertyElseAddedId, ResponsePropertyElseRemovedId, propName, responseStatus, appendPropResultItem)
						})
				}
			}
		}
	}
	return result
}

func checkConditionalSchemaDiffResponse(schemaDiff *diff.SchemaDiff, addedId, removedId, responseStatus string, appendResultItem func(string, ...any)) {
	if schemaDiff == nil {
		return
	}
	if schemaDiff.SchemaAdded {
		appendResultItem(addedId, responseStatus)
	}
	if schemaDiff.SchemaDeleted {
		appendResultItem(removedId, responseStatus)
	}
}

func checkConditionalSchemaDiffPropertyResponse(schemaDiff *diff.SchemaDiff, addedId, removedId, propName, responseStatus string, appendResultItem func(string, ...any)) {
	if schemaDiff == nil {
		return
	}
	if schemaDiff.SchemaAdded {
		appendResultItem(addedId, propName, responseStatus)
	}
	if schemaDiff.SchemaDeleted {
		appendResultItem(removedId, propName, responseStatus)
	}
}
