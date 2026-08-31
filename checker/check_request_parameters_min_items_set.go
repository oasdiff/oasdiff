package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinItemsSetId = "request-parameter-min-items-set"
)

func RequestParameterMinItemsSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		_, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "minItems")
		minItemsDiff := p.paramDiff.SchemaDiff.MinItemsDiff
		if minItemsDiff == nil {
			return
		}
		if !uintBoundSet(minItemsDiff) {
			return
		}

		result = append(result, p.opInfo.NewApiChange(
			RequestParameterMinItemsSetId,
			[]any{p.location, p.name, minItemsDiff.To},
			commentId(RequestParameterMinItemsSetId),
		).WithSources(nil, revisionSource))
	})
	return result
}
