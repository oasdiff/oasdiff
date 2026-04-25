package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyContentSchemaAddedId      = "response-body-content-schema-added"
	ResponseBodyContentSchemaRemovedId    = "response-body-content-schema-removed"
	ResponseBodyContentMediaTypeChangedId = "response-body-content-media-type-changed"
	ResponseBodyContentEncodingChangedId  = "response-body-content-encoding-changed"

	ResponsePropertyContentSchemaAddedId      = "response-property-content-schema-added"
	ResponsePropertyContentSchemaRemovedId    = "response-property-content-schema-removed"
	ResponsePropertyContentMediaTypeChangedId = "response-property-content-media-type-changed"
	ResponsePropertyContentEncodingChangedId  = "response-property-content-encoding-changed"
)

func ResponsePropertyContentUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					if mediaTypeDiff.SchemaDiff.ContentSchemaDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contentSchema")
						if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaAdded {
							result = append(result, NewApiChange(
								ResponseBodyContentSchemaAddedId,
								config,
								[]any{responseStatus},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
						}
						if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaDeleted {
							result = append(result, NewApiChange(
								ResponseBodyContentSchemaRemovedId,
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

					if mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contentMediaType")
						d := mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff
						result = append(result, NewApiChange(
							ResponseBodyContentMediaTypeChangedId,
							config,
							[]any{d.From, d.To, responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}

					if mediaTypeDiff.SchemaDiff.ContentEncodingDiff != nil {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contentEncoding")
						d := mediaTypeDiff.SchemaDiff.ContentEncodingDiff
						result = append(result, NewApiChange(
							ResponseBodyContentEncodingChangedId,
							config,
							[]any{d.From, d.To, responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							propName := propertyFullName(propertyPath, propertyName)

							if propertyDiff.ContentSchemaDiff != nil {
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "contentSchema")
								if propertyDiff.ContentSchemaDiff.SchemaAdded {
									result = append(result, NewApiChange(
										ResponsePropertyContentSchemaAddedId,
										config,
										[]any{propName, responseStatus},
										"",
										operationsSources,
										operationItem.Revision,
										operation,
										path,
									).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
								}
								if propertyDiff.ContentSchemaDiff.SchemaDeleted {
									result = append(result, NewApiChange(
										ResponsePropertyContentSchemaRemovedId,
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

							if propertyDiff.ContentMediaTypeDiff != nil {
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "contentMediaType")
								d := propertyDiff.ContentMediaTypeDiff
								result = append(result, NewApiChange(
									ResponsePropertyContentMediaTypeChangedId,
									config,
									[]any{propName, d.From, d.To, responseStatus},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(propBaseSource, propRevisionSource).WithDetails(mediaTypeDetails))
							}

							if propertyDiff.ContentEncodingDiff != nil {
								propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "contentEncoding")
								d := propertyDiff.ContentEncodingDiff
								result = append(result, NewApiChange(
									ResponsePropertyContentEncodingChangedId,
									config,
									[]any{propName, d.From, d.To, responseStatus},
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
