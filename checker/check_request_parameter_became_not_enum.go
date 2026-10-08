package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecameNotEnumId         = "request-parameter-became-not-enum"
	RequestParameterPropertyBecameNotEnumId = "request-parameter-property-became-not-enum"
)

func RequestParameterBecameNotEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if enumRemoved(p.paramDiff.SchemaDiff) {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "enum")
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterBecameNotEnumId,
				[]any{p.location, p.name},
				"",
			).WithSchema(p.paramDiff.SchemaDiff, p.paramDiff.SchemaDiff, "").WithSources(baseSource, revisionSource))
		}

		schemawalk.ModifiedProperties(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				if !enumRemoved(propertyDiff) {
					return
				}
				baseSource, revisionSource := location.SchemaFieldSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "enum")
				result = append(result, p.opInfo.NewApiChange(
					RequestParameterPropertyBecameNotEnumId,
					[]any{schemawalk.PropertyFullName(propertyPath, propertyName), p.location, p.name},
					"",
				).WithSchema(p.paramDiff.SchemaDiff, propertyDiff, schemawalk.PropertyFullName(propertyPath, propertyName)).WithSources(baseSource, revisionSource))
			})
	})
	return result
}
