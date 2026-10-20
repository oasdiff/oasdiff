package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// VariablesDiff describes the changes between a pair of sets of server variable objects: https://swagger.io/specification/#server-variable-object
type VariablesDiff struct {
	Added    []string          `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string          `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedVariables `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedVariables is map of variable names to their respective diffs
type ModifiedVariables map[string]*VariableDiff

// Empty indicates whether a change was found in this element
func (diff *VariablesDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newVariablesDiff() *VariablesDiff {
	return &VariablesDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedVariables{},
	}
}

func getVariablesDiff(config *Config, variables1, variables2 map[string]*openapi3.ServerVariable) (*VariablesDiff, error) {
	diff, err := getVariablesDiffInternal(config, variables1, variables2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getVariablesDiffInternal(config *Config, variables1, variables2 map[string]*openapi3.ServerVariable) (*VariablesDiff, error) {
	result := newVariablesDiff()

	added, deleted, common := keysDiff(variables1, variables2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		diff, err := getVariableDiff(config, variables1[name], variables2[name])
		if err != nil {
			return nil, err
		}
		if !diff.Empty() {
			result.Modified[name] = diff
		}
	}

	return result, nil
}
