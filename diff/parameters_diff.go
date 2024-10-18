package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

// ParametersDiff describes the changes between a pair of lists of parameter objects: https://swagger.io/specification/#parameter-object
type ParametersDiff struct {
	Added    utils.StringList `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ParamDiffs       `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *ParametersDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newParametersDiff() *ParametersDiff {
	return &ParametersDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ParamDiffs{},
	}
}

func getParametersDiff(config *Config, state *state, params1, params2 openapi3.ParametersMap) (*ParametersDiff, error) {
	diff, err := getParametersDiffInternal(config, state, params1, params2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getParametersDiffInternal(config *Config, state *state, params1, params2 openapi3.ParametersMap) (*ParametersDiff, error) {

	result := newParametersDiff()

	for paramName1, paramRef1 := range params1 {

		if paramRef2, ok := params2[paramName1]; ok {

			param1, err := derefParam(paramRef1)
			if err != nil {
				return nil, err
			}

			param2, err := derefParam(paramRef2)
			if err != nil {
				return nil, err
			}

			diff, err := getParameterDiff(config, state, param1, param2)
			if err != nil {
				return nil, err
			}

			if !diff.Empty() {
				result.Modified[paramName1] = diff
			}
		} else {
			result.Deleted = append(result.Deleted, paramName1)
		}
	}

	for paramName2 := range params2 {
		if _, ok := params1[paramName2]; !ok {
			result.Added = append(result.Added, paramName2)
		}
	}

	return result, nil
}

func (diff *ParametersDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
