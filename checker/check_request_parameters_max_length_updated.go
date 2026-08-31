package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxLengthDecreasedId = "request-parameter-max-length-decreased"
	RequestParameterMaxLengthIncreasedId = "request-parameter-max-length-increased"
)

func RequestParameterMaxLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		maxLengthDiff := p.paramDiff.SchemaDiff.MaxLengthDiff
		if maxLengthDiff == nil {
			return
		}
		if maxLengthDiff.From == nil ||
			maxLengthDiff.To == nil {
			return
		}

		id := RequestParameterMaxLengthDecreasedId
		if !isDecreasedValue(maxLengthDiff) {
			id = RequestParameterMaxLengthIncreasedId
		}

		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "maxLength")
		result = append(result, p.opInfo.NewApiChange(
			id,
			[]any{p.location, p.name, maxLengthDiff.From, maxLengthDiff.To},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
