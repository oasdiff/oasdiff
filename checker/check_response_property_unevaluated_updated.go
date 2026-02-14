package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyUnevaluatedItemsAddedId        = "response-body-unevaluated-items-added"
	ResponseBodyUnevaluatedItemsRemovedId      = "response-body-unevaluated-items-removed"
	ResponseBodyUnevaluatedPropertiesAddedId   = "response-body-unevaluated-properties-added"
	ResponseBodyUnevaluatedPropertiesRemovedId = "response-body-unevaluated-properties-removed"

	ResponsePropertyUnevaluatedItemsAddedId        = "response-property-unevaluated-items-added"
	ResponsePropertyUnevaluatedItemsRemovedId      = "response-property-unevaluated-items-removed"
	ResponsePropertyUnevaluatedPropertiesAddedId   = "response-property-unevaluated-properties-added"
	ResponsePropertyUnevaluatedPropertiesRemovedId = "response-property-unevaluated-properties-removed"
)

func ResponsePropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff != nil {
						if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaAdded {
							appendResultItem(ResponseBodyUnevaluatedItemsAddedId, responseStatus)
						}
						if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
							appendResultItem(ResponseBodyUnevaluatedItemsRemovedId, responseStatus)
						}
					}

					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff != nil {
						if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
							appendResultItem(ResponseBodyUnevaluatedPropertiesAddedId, responseStatus)
						}
						if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
							appendResultItem(ResponseBodyUnevaluatedPropertiesRemovedId, responseStatus)
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							propName := propertyFullName(propertyPath, propertyName)

							if propertyDiff.UnevaluatedItemsDiff != nil {
								if propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
									appendResultItem(ResponsePropertyUnevaluatedItemsAddedId, propName, responseStatus)
								}
								if propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
									appendResultItem(ResponsePropertyUnevaluatedItemsRemovedId, propName, responseStatus)
								}
							}

							if propertyDiff.UnevaluatedPropertiesDiff != nil {
								if propertyDiff.UnevaluatedPropertiesDiff.SchemaAdded {
									appendResultItem(ResponsePropertyUnevaluatedPropertiesAddedId, propName, responseStatus)
								}
								if propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
									appendResultItem(ResponsePropertyUnevaluatedPropertiesRemovedId, propName, responseStatus)
								}
							}
						})
				}
			}
		}
	}
	return result
}
