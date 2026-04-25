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
					if mediaTypeDiff.SchemaDiff != nil && mediaTypeDiff.SchemaDiff.ConstDiff != nil {
						constDiff := mediaTypeDiff.SchemaDiff.ConstDiff
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "const")

						if constDiff.From == nil {
							result = append(result, NewApiChange(
								ResponseBodyConstAddedId,
								config,
								[]any{mediaType, constDiff.To, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						} else if constDiff.To == nil {
							result = append(result, NewApiChange(
								ResponseBodyConstRemovedId,
								config,
								[]any{mediaType, constDiff.From, responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, nil).WithDetails(mediaTypeDetails))
						} else {
							result = append(result, NewApiChange(
								ResponseBodyConstChangedId,
								config,
								[]any{mediaType, constDiff.From, constDiff.To, responseStatus},
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
							if propertyDiff == nil || propertyDiff.Revision == nil || propertyDiff.ConstDiff == nil {
								return
							}

							constDiff := propertyDiff.ConstDiff
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "const")

							if constDiff.From == nil {
								result = append(result, NewApiChange(
									ResponsePropertyConstAddedId,
									config,
									[]any{propertyName, constDiff.To, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							} else if constDiff.To == nil {
								result = append(result, NewApiChange(
									ResponsePropertyConstRemovedId,
									config,
									[]any{propertyName, constDiff.From, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, nil).WithDetails(mediaTypeDetails))
							} else {
								result = append(result, NewApiChange(
									ResponsePropertyConstChangedId,
									config,
									[]any{propertyName, constDiff.From, constDiff.To, responseStatus},
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
	}
	return result
}
