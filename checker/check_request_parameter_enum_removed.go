package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterEnumRemovedId         = "request-parameter-enum-removed"
	RequestParameterPropertyEnumRemovedId = "request-parameter-property-enum-removed"
)

func RequestParameterEnumRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if enumRemoved(p.paramDiff.SchemaDiff) {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "enum")
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterEnumRemovedId,
				[]any{p.location, p.name},
				"",
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))
		}

		schemawalk.ModifiedProperties(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				if !enumRemoved(propertyDiff) {
					return
				}
				baseSource, revisionSource := location.SchemaFieldSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "enum")
				result = append(result, p.opInfo.NewApiChange(
					RequestParameterPropertyEnumRemovedId,
					[]any{schemawalk.PropertyFullName(propertyPath, propertyName), p.location, p.name},
					"",
				).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))
			})
	})
	return result
}
