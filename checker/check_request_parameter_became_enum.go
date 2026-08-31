package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecameEnumId = "request-parameter-became-enum"
)

func RequestParameterBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "enum")

		if enumDiff := p.paramDiff.SchemaDiff.EnumDiff; enumDiff == nil || !enumDiff.EnumAdded {
			return
		}

		result = append(result, p.opInfo.NewApiChange(
			RequestParameterBecameEnumId,
			[]any{p.location, p.name},
			"",
		).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))
	})
	return result
}
