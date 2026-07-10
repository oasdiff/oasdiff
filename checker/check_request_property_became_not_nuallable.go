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
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
		switch nullabilityChange(info.schemaDiff) {
		case becameNullable:
			result = append(result, info.newChange(RequestBodyBecomeNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		case becameNotNullable:
			result = append(result, info.newChange(RequestBodyBecomeNotNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
			switch nullabilityChange(p.propertyDiff) {
			case becameNullable:
				result = append(result, p.newChange(RequestPropertyBecomeNullableId, []any{propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			case becameNotNullable:
				result = append(result, p.newChange(RequestPropertyBecomeNotNullableId, []any{propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
