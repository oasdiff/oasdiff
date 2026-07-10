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
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
		switch nullabilityChange(info.schemaDiff) {
		case becameNullable:
			result = append(result, info.newChange(ResponseBodyBecameNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		case becameNotNullable:
			result = append(result, info.newChange(ResponseBodyBecameNotNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
			switch nullabilityChange(p.propertyDiff) {
			case becameNullable:
				result = append(result, p.newChange(
					ResponsePropertyBecameNullableId,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			case becameNotNullable:
				result = append(result, p.newChange(
					ResponsePropertyBecameNotNullableId,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
