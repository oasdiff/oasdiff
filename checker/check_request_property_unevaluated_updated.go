package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyUnevaluatedItemsAddedId        = "request-body-unevaluated-items-added"
	RequestBodyUnevaluatedItemsRemovedId      = "request-body-unevaluated-items-removed"
	RequestBodyUnevaluatedPropertiesAddedId   = "request-body-unevaluated-properties-added"
	RequestBodyUnevaluatedPropertiesRemovedId = "request-body-unevaluated-properties-removed"

	RequestPropertyUnevaluatedItemsAddedId        = "request-property-unevaluated-items-added"
	RequestPropertyUnevaluatedItemsRemovedId      = "request-property-unevaluated-items-removed"
	RequestPropertyUnevaluatedPropertiesAddedId   = "request-property-unevaluated-properties-added"
	RequestPropertyUnevaluatedPropertiesRemovedId = "request-property-unevaluated-properties-removed"
)

func RequestPropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff != nil {
					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaAdded {
						appendResultItem(RequestBodyUnevaluatedItemsAddedId)
					}
					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
						appendResultItem(RequestBodyUnevaluatedItemsRemovedId)
					}
				}

				if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff != nil {
					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
						appendResultItem(RequestBodyUnevaluatedPropertiesAddedId)
					}
					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
						appendResultItem(RequestBodyUnevaluatedPropertiesRemovedId)
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						propName := propertyFullName(propertyPath, propertyName)

						if propertyDiff.UnevaluatedItemsDiff != nil {
							if propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
								appendResultItem(RequestPropertyUnevaluatedItemsAddedId, propName)
							}
							if propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
								appendResultItem(RequestPropertyUnevaluatedItemsRemovedId, propName)
							}
						}

						if propertyDiff.UnevaluatedPropertiesDiff != nil {
							if propertyDiff.UnevaluatedPropertiesDiff.SchemaAdded {
								appendResultItem(RequestPropertyUnevaluatedPropertiesAddedId, propName)
							}
							if propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
								appendResultItem(RequestPropertyUnevaluatedPropertiesRemovedId, propName)
							}
						}
					})
			}
		}
	}
	return result
}
