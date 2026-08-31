package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinLengthIncreasedId = "request-parameter-min-length-increased"
	RequestParameterMinLengthDecreasedId = "request-parameter-min-length-decreased"
)

func RequestParameterMinLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "minLength")
		minLengthDiff := p.paramDiff.SchemaDiff.MinLengthDiff
		if minLengthDiff == nil {
			return
		}
		if minLengthDiff.From == nil ||
			minLengthDiff.To == nil {
			return
		}

		id := RequestParameterMinLengthIncreasedId
		if isDecreasedValue(minLengthDiff) {
			id = RequestParameterMinLengthDecreasedId
		}

		result = append(result, p.opInfo.NewApiChange(
			id,
			[]any{p.location, p.name, minLengthDiff.From, minLengthDiff.To},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
