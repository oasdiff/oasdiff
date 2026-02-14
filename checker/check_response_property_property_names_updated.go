package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPropertyNamesAddedId       = "response-body-property-names-added"
	ResponseBodyPropertyNamesRemovedId     = "response-body-property-names-removed"
	ResponsePropertyPropertyNamesAddedId   = "response-property-property-names-added"
	ResponsePropertyPropertyNamesRemovedId = "response-property-property-names-removed"
)

func ResponsePropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff != nil {
						if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaAdded {
							appendResultItem(ResponseBodyPropertyNamesAddedId, responseStatus)
						}
						if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaDeleted {
							appendResultItem(ResponseBodyPropertyNamesRemovedId, responseStatus)
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
								appendResultItem(ResponsePropertyPropertyNamesAddedId, propName, responseStatus)
							}
							if propertyDiff.PropertyNamesDiff.SchemaDeleted {
								appendResultItem(ResponsePropertyPropertyNamesRemovedId, propName, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
