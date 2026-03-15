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

					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "unevaluatedItems")
						if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaAdded {
							result = append(result, NewApiChange(
								ResponseBodyUnevaluatedItemsAddedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								ResponseBodyUnevaluatedItemsRemovedId,
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

					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "unevaluatedProperties")
						if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
							result = append(result, NewApiChange(
								ResponseBodyUnevaluatedPropertiesAddedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								ResponseBodyUnevaluatedPropertiesRemovedId,
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
							propName := propertyFullName(propertyPath, propertyName)

							if propertyDiff.UnevaluatedItemsDiff != nil {
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "unevaluatedItems")
								if propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
									result = append(result, NewApiChange(
										ResponsePropertyUnevaluatedItemsAddedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
									result = append(result, NewApiChange(
										ResponsePropertyUnevaluatedItemsRemovedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
								}
							}

							if propertyDiff.UnevaluatedPropertiesDiff != nil {
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "unevaluatedProperties")
								if propertyDiff.UnevaluatedPropertiesDiff.SchemaAdded {
									result = append(result, NewApiChange(
										ResponsePropertyUnevaluatedPropertiesAddedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
									result = append(result, NewApiChange(
										ResponsePropertyUnevaluatedPropertiesRemovedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
								}
							}
						})
				}
			}
		}
	}
	return result
}
