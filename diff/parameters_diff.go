package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ParametersDiff describes the changes between a pair of lists of parameter objects: https://swagger.io/specification/#parameter-object
type ParametersDiff struct {
	Added    []string   `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string   `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ParamDiffs `json:"modified,omitempty" yaml:"modified,omitempty"`
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
		Added:    []string{},
		Deleted:  []string{},
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

	added, deleted, common := keysDiff(params1, params2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		param1, err := derefParam(params1[name])
		if err != nil {
			return nil, err
		}

		param2, err := derefParam(params2[name])
		if err != nil {
			return nil, err
		}

		diff, err := getParameterDiff(config, state, param1, param2)
		if err != nil {
			return nil, err
		}

		if !diff.Empty() {
			result.Modified[name] = diff
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
