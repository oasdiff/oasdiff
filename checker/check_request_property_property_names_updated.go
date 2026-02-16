package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPropertyNamesAddedId       = "request-body-property-names-added"
	RequestBodyPropertyNamesRemovedId     = "request-body-property-names-removed"
	RequestPropertyPropertyNamesAddedId   = "request-property-property-names-added"
	RequestPropertyPropertyNamesRemovedId = "request-property-property-names-removed"
)

func RequestPropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.PropertyNamesDiff != nil {
					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaAdded {
						appendResultItem(RequestBodyPropertyNamesAddedId)
					}
					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaDeleted {
						appendResultItem(RequestBodyPropertyNamesRemovedId)
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.PropertyNamesDiff == nil {
							return
						}
						propName := propertyFullName(propertyPath, propertyName)
						if propertyDiff.PropertyNamesDiff.SchemaAdded {
							appendResultItem(RequestPropertyPropertyNamesAddedId, propName)
						}
						if propertyDiff.PropertyNamesDiff.SchemaDeleted {
							appendResultItem(RequestPropertyPropertyNamesRemovedId, propName)
						}
					})
			}
		}
	}
	return result
}
