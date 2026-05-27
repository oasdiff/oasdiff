package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameNullableId = "response-property-became-nullable"
	ResponseBodyBecameNullableId     = "response-body-became-nullable"
)

func ResponsePropertyBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
		if info.schemaDiff.NullableDiff != nil && info.schemaDiff.NullableDiff.To == true {
			result = append(result, info.newChange(ResponseBodyBecameNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		} else if nullAddedToTypeArray(info.schemaDiff.TypeDiff) {
			// OpenAPI 3.1: type changed from "string" to ["string", "null"]
			result = append(result, info.newChange(ResponseBodyBecameNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
			nullableDiff := p.propertyDiff.NullableDiff
			if nullableDiff != nil {
				if nullableDiff.To != true {
					return
				}
				result = append(result, p.newChange(
					ResponsePropertyBecameNullableId,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}
			// OpenAPI 3.1: type changed from "string" to ["string", "null"]
			if nullAddedToTypeArray(p.propertyDiff.TypeDiff) {
				result = append(result, p.newChange(
					ResponsePropertyBecameNullableId,
					[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
