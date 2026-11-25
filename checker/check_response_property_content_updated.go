package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
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

func ResponsePropertyContentUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		result = append(result, bodySubschemaKeywordChanges(info, DirectionResponse, contentSchemaKeyword)...)

		if info.schemaDiff.ContentMediaTypeDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentMediaType")
			d := info.schemaDiff.ContentMediaTypeDiff
			result = append(result, info.newChange(ResponseBodyContentMediaTypeChangedId, []any{d.From, d.To, info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}

		if info.schemaDiff.ContentEncodingDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "contentEncoding")
			d := info.schemaDiff.ContentEncodingDiff
			result = append(result, info.newChange(ResponseBodyContentEncodingChangedId, []any{d.From, d.To, info.responseStatus}, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)

			result = append(result, propertySubschemaKeywordChanges(p, DirectionResponse, contentSchemaKeyword)...)

			if p.propertyDiff.ContentMediaTypeDiff != nil {
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentMediaType")
				d := p.propertyDiff.ContentMediaTypeDiff
				result = append(result, p.newChange(ResponsePropertyContentMediaTypeChangedId, []any{propName, d.From, d.To, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}

			if p.propertyDiff.ContentEncodingDiff != nil {
				propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "contentEncoding")
				d := p.propertyDiff.ContentEncodingDiff
				result = append(result, p.newChange(ResponsePropertyContentEncodingChangedId, []any{propName, d.From, d.To, info.responseStatus}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
