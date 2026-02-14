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
						appendResultItem(RequestBodyContentSchemaAddedId)
					}
					if mediaTypeDiff.SchemaDiff.ContentSchemaDiff.SchemaDeleted {
						appendResultItem(RequestBodyContentSchemaRemovedId)
					}
				}

				if mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff != nil {
					appendResultItem(RequestBodyContentMediaTypeChangedId, mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff.From, mediaTypeDiff.SchemaDiff.ContentMediaTypeDiff.To)
				}

				if mediaTypeDiff.SchemaDiff.ContentEncodingDiff != nil {
					appendResultItem(RequestBodyContentEncodingChangedId, mediaTypeDiff.SchemaDiff.ContentEncodingDiff.From, mediaTypeDiff.SchemaDiff.ContentEncodingDiff.To)
				}

				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						propName := propertyFullName(propertyPath, propertyName)

						if propertyDiff.ContentSchemaDiff != nil {
							if propertyDiff.ContentSchemaDiff.SchemaAdded {
								appendResultItem(RequestPropertyContentSchemaAddedId, propName)
							}
							if propertyDiff.ContentSchemaDiff.SchemaDeleted {
								appendResultItem(RequestPropertyContentSchemaRemovedId, propName)
							}
						}

						if propertyDiff.ContentMediaTypeDiff != nil {
							appendResultItem(RequestPropertyContentMediaTypeChangedId, propName, propertyDiff.ContentMediaTypeDiff.From, propertyDiff.ContentMediaTypeDiff.To)
						}

						if propertyDiff.ContentEncodingDiff != nil {
							appendResultItem(RequestPropertyContentEncodingChangedId, propName, propertyDiff.ContentEncodingDiff.From, propertyDiff.ContentEncodingDiff.To)
						}
					})
			}
		}
	}
	return result
}
