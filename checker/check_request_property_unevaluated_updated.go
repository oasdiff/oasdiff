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

				if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "unevaluatedItems")
					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaAdded {
						result = append(result, NewApiChange(
							RequestBodyUnevaluatedItemsAddedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					if mediaTypeDiff.SchemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
						result = append(result, NewApiChange(
							RequestBodyUnevaluatedItemsRemovedId,
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

				if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "unevaluatedProperties")
					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
						result = append(result, NewApiChange(
							RequestBodyUnevaluatedPropertiesAddedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					if mediaTypeDiff.SchemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
						result = append(result, NewApiChange(
							RequestBodyUnevaluatedPropertiesRemovedId,
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
						propName := propertyFullName(propertyPath, propertyName)

						if propertyDiff.UnevaluatedItemsDiff != nil {
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "unevaluatedItems")
							if propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
								result = append(result, NewApiChange(
									RequestPropertyUnevaluatedItemsAddedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
								result = append(result, NewApiChange(
									RequestPropertyUnevaluatedItemsRemovedId,
									config,
									[]any{propName},
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
									RequestPropertyUnevaluatedPropertiesAddedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
								result = append(result, NewApiChange(
									RequestPropertyUnevaluatedPropertiesRemovedId,
									config,
									[]any{propName},
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
	return result
}
