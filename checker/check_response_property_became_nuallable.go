package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameNullableId    = "response-property-became-nullable"
	ResponseBodyBecameNullableId        = "response-body-became-nullable"
	ResponsePropertyBecameNotNullableId = "response-property-became-not-nullable"
	ResponseBodyBecameNotNullableId     = "response-body-became-not-nullable"
)

func ResponsePropertyBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if id := nullabilityChangeId(info.schemaDiff, ResponseBodyBecameNullableId, ResponseBodyBecameNotNullableId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
			result = append(result, info.newChange(id, nil, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			if id := nullabilityChangeId(p.propertyDiff, ResponsePropertyBecameNullableId, ResponsePropertyBecameNotNullableId); id != "" {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
