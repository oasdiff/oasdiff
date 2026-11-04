package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	NewRequiredRequestDefaultParameterToExistingPathId = "new-required-request-default-parameter-to-existing-path"
	NewOptionalRequestDefaultParameterToExistingPathId = "new-optional-request-default-parameter-to-existing-path"
)

// NewRequestNonPathDefaultParameterCheck reports a parameter added to a path
// item at each operation that didn't have it before.
func NewRequestNonPathDefaultParameterCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkOperationsParameters(diffReport, operationsSources, config, func(o operationParams) {
		for paramLocation, paramItems := range o.ParametersDiff.Added {
			if paramLocation == "path" {
				continue
			}
			for _, paramName := range paramItems {
				// the operation's own parameter is reported by NewRequestNonPathParameterCheck
				if o.Revision.Parameters.GetByInAndName(paramLocation, paramName) != nil {
					continue
				}
				param := o.RevisionParameters.GetByInAndName(paramLocation, paramName)
				id := NewRequiredRequestDefaultParameterToExistingPathId
				if !param.Required {
					id = NewOptionalRequestDefaultParameterToExistingPathId
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
