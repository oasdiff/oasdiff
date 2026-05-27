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

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.ContentSchemaDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentSchema")
			if info.schemaDiff.ContentSchemaDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyContentSchemaAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.ContentSchemaDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyContentSchemaRemovedId, nil, "").
					WithSources(baseSource, nil))
			}
		}

		if info.schemaDiff.ContentMediaTypeDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentMediaType")
			d := info.schemaDiff.ContentMediaTypeDiff
			result = append(result, info.newChange(RequestBodyContentMediaTypeChangedId, []any{d.From, d.To}, "").
				WithSources(baseSource, revisionSource))
		}

		if info.schemaDiff.ContentEncodingDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentEncoding")
			d := info.schemaDiff.ContentEncodingDiff
			result = append(result, info.newChange(RequestBodyContentEncodingChangedId, []any{d.From, d.To}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if p.propertyDiff.ContentSchemaDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentSchema")
				if p.propertyDiff.ContentSchemaDiff.SchemaAdded {
					result = append(result, p.newChange(RequestPropertyContentSchemaAddedId, []any{propName}, "").
						WithSources(nil, propRevisionSource))
				}
				if p.propertyDiff.ContentSchemaDiff.SchemaDeleted {
					result = append(result, p.newChange(RequestPropertyContentSchemaRemovedId, []any{propName}, "").
						WithSources(propBaseSource, nil))
				}
			}

			if p.propertyDiff.ContentMediaTypeDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentMediaType")
				d := p.propertyDiff.ContentMediaTypeDiff
				result = append(result, p.newChange(RequestPropertyContentMediaTypeChangedId, []any{propName, d.From, d.To}, "").
					WithSources(propBaseSource, propRevisionSource))
			}

			if p.propertyDiff.ContentEncodingDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentEncoding")
				d := p.propertyDiff.ContentEncodingDiff
				result = append(result, p.newChange(RequestPropertyContentEncodingChangedId, []any{propName, d.From, d.To}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
