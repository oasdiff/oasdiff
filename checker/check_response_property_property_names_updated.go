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

					if mediaTypeDiff.SchemaDiff.PropertyNamesDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "propertyNames")
						if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaAdded {
							result = append(result, NewApiChange(
								ResponseBodyPropertyNamesAddedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if mediaTypeDiff.SchemaDiff.PropertyNamesDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								ResponseBodyPropertyNamesRemovedId,
								config,
								[]any{responseStatus},
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
									ResponsePropertyPropertyNamesAddedId,
									config,
									[]any{propName, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if propertyDiff.PropertyNamesDiff.SchemaDeleted {
								result = append(result, NewApiChange(
									ResponsePropertyPropertyNamesRemovedId,
									config,
									[]any{propName, responseStatus},
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
	}
	return result
}
