package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyConstAddedId       = "response-body-const-added"
	ResponseBodyConstRemovedId     = "response-body-const-removed"
	ResponseBodyConstChangedId     = "response-body-const-changed"
	ResponsePropertyConstAddedId   = "response-property-const-added"
	ResponsePropertyConstRemovedId = "response-property-const-removed"
	ResponsePropertyConstChangedId = "response-property-const-changed"
)

func ResponsePropertyConstChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
				if responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}

				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for mediaType, mediaTypeDiff := range modifiedMediaTypes {
					mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
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
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.ConstDiff != nil {
						constDiff := mediaTypeDiff.SchemaDiff.ConstDiff
						if constDiff.From == nil {
							appendResultItem(ResponseBodyConstAddedId, mediaType, constDiff.To, responseStatus)
						} else if constDiff.To == nil {
							appendResultItem(ResponseBodyConstRemovedId, mediaType, constDiff.From, responseStatus)
						} else {
							appendResultItem(ResponseBodyConstChangedId, mediaType, constDiff.From, constDiff.To, responseStatus)
						}
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff == nil || propertyDiff.Revision == nil || propertyDiff.ConstDiff == nil {
								return
							}

							constDiff := propertyDiff.ConstDiff
							if constDiff.From == nil {
								appendResultItem(ResponsePropertyConstAddedId, propertyName, constDiff.To, responseStatus)
							} else if constDiff.To == nil {
								appendResultItem(ResponsePropertyConstRemovedId, propertyName, constDiff.From, responseStatus)
							} else {
								appendResultItem(ResponsePropertyConstChangedId, propertyName, constDiff.From, constDiff.To, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
