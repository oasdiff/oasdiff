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
				if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.ConstDiff != nil {
					constDiff := mediaTypeDiff.SchemaDiff.ConstDiff
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "const")

					if constDiff.From == nil {
						result = append(result, NewApiChange(
							RequestBodyConstAddedId,
							config,
							[]any{mediaType, constDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					} else if constDiff.To == nil {
						result = append(result, NewApiChange(
							RequestBodyConstRemovedId,
							config,
							[]any{mediaType, constDiff.From},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
					} else {
						result = append(result, NewApiChange(
							RequestBodyConstChangedId,
							config,
							[]any{mediaType, constDiff.From, constDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff == nil || propertyDiff.ConstDiff == nil {
							return
						}

						constDiff := propertyDiff.ConstDiff
						propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "const")

						if constDiff.From == nil {
							result = append(result, NewApiChange(
								RequestPropertyConstAddedId,
								config,
								[]any{propertyName, constDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
						} else if constDiff.To == nil {
							result = append(result, NewApiChange(
								RequestPropertyConstRemovedId,
								config,
								[]any{propertyName, constDiff.From},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								RequestPropertyConstChangedId,
								config,
								[]any{propertyName, constDiff.From, constDiff.To},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
