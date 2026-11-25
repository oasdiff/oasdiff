package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
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

func RequestPropertyContentUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, bodySingleSchemaKeywordChanges(info, DirectionRequest, contentSchemaKeyword)...)

		if info.schemaDiff.ContentMediaTypeDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentMediaType")
			d := info.schemaDiff.ContentMediaTypeDiff
			result = append(result, info.newChange(RequestBodyContentMediaTypeChangedId, []any{d.From, d.To}, "").
				WithSources(baseSource, revisionSource))
		}

		if info.schemaDiff.ContentEncodingDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentEncoding")
			d := info.schemaDiff.ContentEncodingDiff
			result = append(result, info.newChange(RequestBodyContentEncodingChangedId, []any{d.From, d.To}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)

			result = append(result, propertySingleSchemaKeywordChanges(p, DirectionRequest, contentSchemaKeyword)...)

			if p.propertyDiff.ContentMediaTypeDiff != nil {
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentMediaType")
				d := p.propertyDiff.ContentMediaTypeDiff
				result = append(result, p.newChange(RequestPropertyContentMediaTypeChangedId, []any{propName, d.From, d.To}, "").
					WithSources(propBaseSource, propRevisionSource))
			}

			if p.propertyDiff.ContentEncodingDiff != nil {
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentEncoding")
				d := p.propertyDiff.ContentEncodingDiff
				result = append(result, p.newChange(RequestPropertyContentEncodingChangedId, []any{propName, d.From, d.To}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
