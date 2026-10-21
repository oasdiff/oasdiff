package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	NewRequiredRequestParameterId = "new-required-request-parameter"
	NewOptionalRequestParameterId = "new-optional-request-parameter"
)

func NewRequestNonPathParameterCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkOperationsParameters(diffReport, operationsSources, config, func(o operationParams) {
		for paramLocation, paramItems := range o.ParametersDiff.Added {
			if paramLocation == "path" {
				// it is processed in the separate check NewRequestPathParameterCheck
				continue
			}
			for _, paramName := range paramItems {
				// a parameter added to the path item is reported by NewRequestNonPathDefaultParameterCheck
				param := o.Revision.Parameters.GetByInAndName(paramLocation, paramName)
				if param == nil {
					continue
				}
				id := NewRequiredRequestParameterId
				if !param.Required {
					id = NewOptionalRequestParameterId
				}
				revisionSource := location.ParameterSource(operationsSources, o.Revision, param)
				result = append(result, o.opInfo.NewApiChange(
					id,
					[]any{paramLocation, paramName},
					"",
				).WithSources(nil, revisionSource))
			}
		}
	})
	return result
}
