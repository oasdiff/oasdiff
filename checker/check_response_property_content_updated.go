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

					if mediaTypeDiff.SchemaDiff.ContentSchemaDiff != nil {
						if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaAdded {
							appendResultItem(ResponseBodyContentSchemaAddedId, responseStatus)
						}
						if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaDeleted {
							appendResultItem(ResponseBodyContentSchemaRemovedId, responseStatus)
						}
					}

					if mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff != nil {
						appendResultItem(ResponseBodyContentMediaTypeChangedId, mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff.From, mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff.To, responseStatus)
					}

					if mediaTypeDiff.SchemaDiff.ContentEncodingDiff != nil {
						appendResultItem(ResponseBodyContentEncodingChangedId, mediaTypeDiff.SchemaDiff.ContentEncodingDiff.From, mediaTypeDiff.SchemaDiff.ContentEncodingDiff.To, responseStatus)
					}

					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							propName := propertyFullName(propertyPath, propertyName)

							if propertyDiff.ContentSchemaDiff != nil {
								if propertyDiff.ContentSchemaDiff.SchemaAdded {
									appendResultItem(ResponsePropertyContentSchemaAddedId, propName, responseStatus)
								}
								if propertyDiff.ContentSchemaDiff.SchemaDeleted {
									appendResultItem(ResponsePropertyContentSchemaRemovedId, propName, responseStatus)
								}
							}

							if propertyDiff.ContentMediaTypeDiff != nil {
								appendResultItem(ResponsePropertyContentMediaTypeChangedId, propName, propertyDiff.ContentMediaTypeDiff.From, propertyDiff.ContentMediaTypeDiff.To, responseStatus)
							}

							if propertyDiff.ContentEncodingDiff != nil {
								appendResultItem(ResponsePropertyContentEncodingChangedId, propName, propertyDiff.ContentEncodingDiff.From, propertyDiff.ContentEncodingDiff.To, responseStatus)
							}
						})
				}
			}
		}
	}
	return result
}
