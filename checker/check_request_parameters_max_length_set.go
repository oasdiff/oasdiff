package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxLengthSetId = "request-parameter-max-length-set"
)

func RequestParameterMaxLengthSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		maxLengthDiff := p.paramDiff.SchemaDiff.MaxLengthDiff
		if maxLengthDiff == nil {
			return
		}
		if maxLengthDiff.From != nil ||
			maxLengthDiff.To == nil {
			return
		}

		_, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "maxLength")
		result = append(result, p.opInfo.NewApiChange(
			RequestParameterMaxLengthSetId,
			[]any{p.location, p.name, maxLengthDiff.To},
			commentId(RequestParameterMaxLengthSetId),
		).WithSources(nil, revisionSource))
	})
	return result
}
