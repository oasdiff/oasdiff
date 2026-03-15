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

				if mediaTypeDiff.SchemaDiff.PropertyNamesDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "propertyNames")
					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaAdded {
						result = append(result, NewApiChange(
							RequestBodyPropertyNamesAddedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaDeleted {
						result = append(result, NewApiChange(
							RequestBodyPropertyNamesRemovedId,
							config,
							nil,
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
						if propertyDiff.PropertyNamesDiff == nil {
							return
						}
						propName := propertyFullName(propertyPath, propertyName)
						propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "propertyNames")
						if propertyDiff.PropertyNamesDiff.SchemaAdded {
							result = append(result, NewApiChange(
								RequestPropertyPropertyNamesAddedId,
								config,
								[]any{propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
						}
						if propertyDiff.PropertyNamesDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								RequestPropertyPropertyNamesRemovedId,
								config,
								[]any{propName},
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
	return result
}
