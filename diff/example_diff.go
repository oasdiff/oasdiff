package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ExampleDiff describes the changes between a pair of example objects: https://swagger.io/specification/#example-object
type ExampleDiff struct {
	ExtensionsDiff    *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	SummaryDiff       *ValueDiff      `json:"summary,omitempty" yaml:"summary,omitempty"`
	DescriptionDiff   *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	ValueDiff         *ValueDiff      `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValueDiff *ValueDiff      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *ExampleDiff) Empty() bool {
	return diff == nil || *diff == ExampleDiff{}
}

func getExampleDiff(config *Config, value1, value2 *openapi3.Example) (*ExampleDiff, error) {
	diff, err := getExampleDiffInternal(config, value1, value2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getExampleDiffInternal(config *Config, value1, value2 *openapi3.Example) (*ExampleDiff, error) {
	result := ExampleDiff{}
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, value1.Extensions, value2.Extensions)
	if err != nil {
		return nil, err
	}

	result.SummaryDiff = getValueDiffConditional(config.IsExcludeSummary(), value1.Summary, value2.Summary)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), value1.Description, value2.Description)
	result.ValueDiff = getValueDiff(value1.Value, value2.Value)
	result.ExternalValueDiff = getValueDiff(value1.ExternalValue, value2.ExternalValue)

	return &result, nil
}
