package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyConstAddedId       = "request-body-const-added"
	RequestBodyConstRemovedId     = "request-body-const-removed"
	RequestBodyConstChangedId     = "request-body-const-changed"
	RequestPropertyConstAddedId   = "request-property-const-added"
	RequestPropertyConstRemovedId = "request-property-const-removed"
	RequestPropertyConstChangedId = "request-property-const-changed"
)

func RequestPropertyConstChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
						appendResultItem(RequestBodyConstAddedId, mediaType, constDiff.To)
					} else if constDiff.To == nil {
						appendResultItem(RequestBodyConstRemovedId, mediaType, constDiff.From)
					} else {
						appendResultItem(RequestBodyConstChangedId, mediaType, constDiff.From, constDiff.To)
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff == nil || propertyDiff.ConstDiff == nil {
							return
						}

						constDiff := propertyDiff.ConstDiff

						if constDiff.From == nil {
							appendResultItem(RequestPropertyConstAddedId, propertyName, constDiff.To)
						} else if constDiff.To == nil {
							appendResultItem(RequestPropertyConstRemovedId, propertyName, constDiff.From)
						} else {
							appendResultItem(RequestPropertyConstChangedId, propertyName, constDiff.From, constDiff.To)
						}
					})
			}
		}
	}
	return result
}
