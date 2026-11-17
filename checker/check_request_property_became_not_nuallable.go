package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecomeNotNullableId     = "request-body-became-not-nullable"
	RequestBodyBecomeNullableId        = "request-body-became-nullable"
	RequestPropertyBecomeNotNullableId = "request-property-became-not-nullable"
	RequestPropertyBecomeNullableId    = "request-property-became-nullable"
)

func RequestPropertyBecameNotNullableCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id, comment := nullabilityChangeId(nil, info.schemaDiff, RequestBodyBecomeNullableId, RequestBodyBecomeNotNullableId); id != "" {
			baseSource, revisionSource := nullabilitySources(operationsSources, info.operationItem, info.schemaDiff)
			result = append(result, info.newChange(id, nil, comment).
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id, comment := nullabilityChangeId(p.parent, p.propertyDiff, RequestPropertyBecomeNullableId, RequestPropertyBecomeNotNullableId); id != "" {
				propBaseSource, propRevisionSource := nullabilitySources(operationsSources, info.operationItem, p.propertyDiff)
				result = append(result, p.newChange(id, []any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)}, comment).
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
