package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecameNullableId            = "request-parameter-became-nullable"
	RequestParameterBecameNotNullableId         = "request-parameter-became-not-nullable"
	RequestParameterPropertyBecameNullableId    = "request-parameter-property-became-nullable"
	RequestParameterPropertyBecameNotNullableId = "request-parameter-property-became-not-nullable"
)

func RequestParameterBecameNullableCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	dialects := dialectsOf(diffReport)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if id, comment := nullabilityChangeId(dialects, nil, p.paramDiff.SchemaDiff, RequestParameterBecameNullableId, RequestParameterBecameNotNullableId); id != "" {
			baseSource, revisionSource := nullabilitySources(dialects, operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff)
			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name},
				comment,
			).WithSchema(p.paramDiff.SchemaDiff, p.paramDiff.SchemaDiff, "").WithSources(baseSource, revisionSource))
		}

		schemawalk.ModifiedProperties(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				if propertyDiff == nil || propertyDiff.Base == nil || propertyDiff.Revision == nil {
					return
				}
				if id, comment := nullabilityChangeId(dialects, parent, propertyDiff, RequestParameterPropertyBecameNullableId, RequestParameterPropertyBecameNotNullableId); id != "" {
					baseSource, revisionSource := nullabilitySources(dialects, operationsSources, p.opInfo.methodDiff, propertyDiff)
					result = append(result, p.opInfo.NewApiChange(
						id,
						[]any{schemawalk.PropertyFullName(propertyPath, propertyName), p.location, p.name},
						comment,
					).WithSchema(p.paramDiff.SchemaDiff, propertyDiff, schemawalk.PropertyFullName(propertyPath, propertyName)).WithSources(baseSource, revisionSource))
				}
			})
	})
	return result
}
