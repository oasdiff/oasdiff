package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// ExamplesDiff describes the changes between a pair of sets of example objects: https://swagger.io/specification/#example-object
type ExamplesDiff struct {
	Added    []string         `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  []string         `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedExamples `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedExamples is map of enconding names to their respective diffs
type ModifiedExamples map[string]*ExampleDiff

// Empty indicates whether a change was found in this element
func (diff *ExamplesDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newExamplesDiff() *ExamplesDiff {
	return &ExamplesDiff{
		Added:    []string{},
		Deleted:  []string{},
		Modified: ModifiedExamples{},
	}
}

func getExamplesDiff(config *Config, examples1, examples2 openapi3.Examples) (*ExamplesDiff, error) {

	diff, err := getExamplesDiffInternal(config, examples1, examples2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getExamplesDiffInternal(config *Config, examples1, examples2 openapi3.Examples) (*ExamplesDiff, error) {

	if config.IsExcludeExamples() {
		return nil, nil
	}

	result := newExamplesDiff()

	added, deleted, common := keysDiff(examples1, examples2)
	result.Added = added
	result.Deleted = deleted

	for _, name := range common {
		value1, err := derefExample(examples1[name])
		if err != nil {
			return nil, err
		}

		value2, err := derefExample(examples2[name])
		if err != nil {
			return nil, err
		}

		diff, err := getExampleDiff(config, value1, value2)
		if err != nil {
			return nil, err
		}

		if !diff.Empty() {
			result.Modified[name] = diff
		}
	}

	return result, nil
}

func derefExample(ref *openapi3.ExampleRef) (*openapi3.Example, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("example reference is nil")
	}

	return ref.Value, nil
}

func (diff *ExamplesDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
