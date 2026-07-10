package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecomeNotNullableId     = "request-body-became-not-nullable"
	RequestBodyBecomeNullableId        = "request-body-became-nullable"
	RequestPropertyBecomeNotNullableId = "request-property-became-not-nullable"
	RequestPropertyBecomeNullableId    = "request-property-became-nullable"
)

func RequestPropertyBecameNotNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id := nullabilityChangeId(info.schemaDiff, RequestBodyBecomeNullableId, RequestBodyBecomeNotNullableId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
			result = append(result, info.newChange(id, nil, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id := nullabilityChangeId(p.propertyDiff, RequestPropertyBecomeNullableId, RequestPropertyBecomeNotNullableId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
				result = append(result, p.newChange(id, []any{propertyFullName(p.propertyPath, p.propertyName)}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
