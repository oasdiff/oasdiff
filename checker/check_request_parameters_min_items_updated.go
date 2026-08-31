package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinItemsIncreasedId = "request-parameter-min-items-increased"
	RequestParameterMinItemsDecreasedId = "request-parameter-min-items-decreased"
)

func RequestParameterMinItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "minItems")
		minItemsDiff := p.paramDiff.SchemaDiff.MinItemsDiff
		if minItemsDiff == nil {
			return
		}
		if uintBoundSet(minItemsDiff) {
			// reported by request-parameter-min-items-set
			return
		}

		id := RequestParameterMinItemsIncreasedId
		if !isIncreasedValue(minItemsDiff) {
			id = RequestParameterMinItemsDecreasedId
		}

		result = append(result, p.opInfo.NewApiChange(
			id,
			[]any{p.location, p.name, minItemsDiff.From, minItemsDiff.To},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
