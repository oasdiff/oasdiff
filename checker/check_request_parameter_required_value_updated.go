package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecomeRequiredId = "request-parameter-became-required"
	RequestParameterBecomeOptionalId = "request-parameter-became-optional"
)

func RequestParameterRequiredValueUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		requiredDiff := p.paramDiff.RequiredDiff
		if requiredDiff == nil {
			return
		}
		baseSource, revisionSource := parameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "required")

		id := RequestParameterBecomeRequiredId

		if requiredDiff.To != true {
			id = RequestParameterBecomeOptionalId
		}

		result = append(result, p.opInfo.NewApiChange(
			id,
			[]any{p.location, p.name},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
