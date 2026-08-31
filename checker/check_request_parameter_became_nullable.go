package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecameNullableId            = "request-parameter-became-nullable"
	RequestParameterBecameNotNullableId         = "request-parameter-became-not-nullable"
	RequestParameterPropertyBecameNullableId    = "request-parameter-property-became-nullable"
	RequestParameterPropertyBecameNotNullableId = "request-parameter-property-became-not-nullable"
)

func RequestParameterBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if id := nullabilityChangeId(p.paramDiff.SchemaDiff, RequestParameterBecameNullableId, RequestParameterBecameNotNullableId); id != "" {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "nullable")
			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name},
				"",
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))
		}

		checkModifiedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				if propertyDiff == nil || propertyDiff.Base == nil || propertyDiff.Revision == nil {
					return
				}
				if id := nullabilityChangeId(propertyDiff, RequestParameterPropertyBecameNullableId, RequestParameterPropertyBecameNotNullableId); id != "" {
					baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "nullable")
					result = append(result, p.opInfo.NewApiChange(
						id,
						[]any{propertyFullName(propertyPath, propertyName), p.location, p.name},
						"",
					).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))
				}
			})
	})
	return result
}
