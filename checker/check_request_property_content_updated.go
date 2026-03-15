package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyContentSchemaAddedId      = "request-body-content-schema-added"
	RequestBodyContentSchemaRemovedId    = "request-body-content-schema-removed"
	RequestBodyContentMediaTypeChangedId = "request-body-content-media-type-changed"
	RequestBodyContentEncodingChangedId  = "request-body-content-encoding-changed"

	RequestPropertyContentSchemaAddedId      = "request-property-content-schema-added"
	RequestPropertyContentSchemaRemovedId    = "request-property-content-schema-removed"
	RequestPropertyContentMediaTypeChangedId = "request-property-content-media-type-changed"
	RequestPropertyContentEncodingChangedId  = "request-property-content-encoding-changed"
)

func RequestPropertyContentUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				if mediaTypeDiff.SchemaDiff.ContentSchemaDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contentSchema")
					if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaAdded {
						result = append(result, NewApiChange(
							RequestBodyContentSchemaAddedId,
							config,
							nil,
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource).WithDetails(mediaTypeDetails))
					}
					if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaDeleted {
						result = append(result, NewApiChange(
							RequestBodyContentSchemaRemovedId,
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

				if mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff != nil {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, mediaTypeDiff.SchemaDiff, "contentMediaType")
					d := mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff
					result = append(result, NewApiChange(
						RequestBodyContentMediaTypeChangedId,
						config,
						[]any{d.From, d.To},
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
						RequestBodyContentEncodingChangedId,
						config,
						[]any{d.From, d.To},
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
									RequestPropertyContentSchemaAddedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									operationItem.Revision,
									operation,
									path,
								).WithSources(nil, propRevisionSource).WithDetails(mediaTypeDetails))
							}
							if propertyDiff.ContentSchemaDiff.SchemaDeleted {
								result = append(result, NewApiChange(
									RequestPropertyContentSchemaRemovedId,
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

						if propertyDiff.ContentMediaTypeDiff != nil {
							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "contentMediaType")
							d := propertyDiff.ContentMediaTypeDiff
							result = append(result, NewApiChange(
								RequestPropertyContentMediaTypeChangedId,
								config,
								[]any{propName, d.From, d.To},
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
								RequestPropertyContentEncodingChangedId,
								config,
								[]any{propName, d.From, d.To},
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
