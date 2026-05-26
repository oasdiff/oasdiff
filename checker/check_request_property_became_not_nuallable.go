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

// Source-attachment behaviour is intentionally non-uniform here, matching the
// pre-migration code:
//   - Body-level NullableDiff emissions carry .WithSources(...).
//   - Body-level OpenAPI-3.1 type-array transitions do not.
//   - Property-level NullableDiff emissions do not.
//   - Property-level OpenAPI-3.1 type-array transitions do.
// This inconsistency predates the walker; preserved as-is to keep the
// migration a pure refactor.
func RequestPropertyBecameNotNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.NullableDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "nullable")
			if info.schemaDiff.NullableDiff.From == true {
				result = append(result, info.newChange(RequestBodyBecomeNotNullableId, nil, "").
					WithSources(baseSource, revisionSource))
			} else if info.schemaDiff.NullableDiff.To == true {
				result = append(result, info.newChange(RequestBodyBecomeNullableId, nil, "").
					WithSources(baseSource, revisionSource))
			}
		} else if nullRemovedFromTypeArray(info.schemaDiff.TypeDiff) {
			// OpenAPI 3.1: type changed from ["string", "null"] to "string"
			result = append(result, info.newChange(RequestBodyBecomeNotNullableId, nil, ""))
		} else if nullAddedToTypeArray(info.schemaDiff.TypeDiff) {
			// OpenAPI 3.1: type changed from "string" to ["string", "null"]
			result = append(result, info.newChange(RequestBodyBecomeNullableId, nil, ""))
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			nullableDiff := p.propertyDiff.NullableDiff
			if nullableDiff != nil {
				if nullableDiff.From == true {
					result = append(result, p.newChange(RequestPropertyBecomeNotNullableId, []any{propName}, ""))
				} else if nullableDiff.To == true {
					result = append(result, p.newChange(RequestPropertyBecomeNullableId, []any{propName}, ""))
				}
			} else if nullRemovedFromTypeArray(p.propertyDiff.TypeDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
				result = append(result, p.newChange(RequestPropertyBecomeNotNullableId, []any{propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			} else if nullAddedToTypeArray(p.propertyDiff.TypeDiff) {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "nullable")
				result = append(result, p.newChange(RequestPropertyBecomeNullableId, []any{propName}, "").
					WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
