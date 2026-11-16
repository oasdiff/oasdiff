package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameNullableId    = "response-property-became-nullable"
	ResponseBodyBecameNullableId        = "response-body-became-nullable"
	ResponsePropertyBecameNotNullableId = "response-property-became-not-nullable"
	ResponseBodyBecameNotNullableId     = "response-body-became-not-nullable"
)

func ResponsePropertyBecameNullableCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	dialects := dialectsOf(diffReport)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id, comment := nullabilityChangeId(dialects, nil, info.schemaDiff, ResponseBodyBecameNullableId, ResponseBodyBecameNotNullableId); id != "" {
			baseSource, revisionSource := nullabilitySources(dialects, operationsSources, info.operationItem, info.schemaDiff)
			result = append(result, info.newChange(id, nil, comment).
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id, comment := nullabilityChangeId(dialects, p.parent, p.propertyDiff, ResponsePropertyBecameNullableId, ResponsePropertyBecameNotNullableId); id != "" {
				propBaseSource, propRevisionSource := nullabilitySources(dialects, operationsSources, info.operationItem, p.propertyDiff)
				result = append(result, p.newChange(
					id,
					[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					comment,
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
