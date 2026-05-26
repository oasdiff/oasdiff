package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameNullableId = "response-property-became-nullable"
	ResponseBodyBecameNullableId     = "response-body-became-nullable"
)

// Mirrors the request-side became-not-nullable shape: NullableDiff emissions
// carry .WithSources(...), OpenAPI-3.1 type-array transitions do not. The
// asymmetry predates the walker; preserved as-is.
func ResponsePropertyBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.NullableDiff != nil && info.schemaDiff.NullableDiff.To == true {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
			result = append(result, info.newChange(ResponseBodyBecameNullableId, nil, "").
				WithSources(baseSource, revisionSource))
		} else if nullAddedToTypeArray(info.schemaDiff.TypeDiff) {
			// OpenAPI 3.1: type changed from "string" to ["string", "null"]
			result = append(result, info.newChange(ResponseBodyBecameNullableId, nil, ""))
		}

		info.walkProperties(func(p propertyInfo) {
			nullableDiff := p.propertyDiff.NullableDiff
			if nullableDiff != nil {
				if nullableDiff.To != true {
					return
				}
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
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
				))
			}
		})
	})

	return result
}
