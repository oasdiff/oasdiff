package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	NewRequestPathParameterId = "new-request-path-parameter"
)

func NewRequestPathParameterCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkOperationsParameters(diffReport, operationsSources, config, func(o operationParams) {
		for _, paramName := range o.ParametersDiff.Added["path"] {
			param := o.RevisionParameters.GetByInAndName("path", paramName)
			revisionSource := location.ParameterSource(operationsSources, o.Revision, param)
			result = append(result, o.opInfo.NewApiChange(
				NewRequestPathParameterId,
				[]any{paramName},
				"",
			).WithSources(nil, revisionSource))
		}
	})
	return result
}
