package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// HeaderDiff describes the changes between a pair of header objects: https://swagger.io/specification/#header-object
type HeaderDiff struct {
	ExtensionsDiff  *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	DescriptionDiff *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	DeprecatedDiff  *ValueDiff      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	RequiredDiff    *ValueDiff      `json:"required,omitempty" yaml:"required,omitempty"`
	ExampleDiff     *ValueDiff      `json:"example,omitempty" yaml:"example,omitempty"`
	ExamplesDiff    *ExamplesDiff   `json:"examples,omitempty" yaml:"examples,omitempty"`
	SchemaDiff      *SchemaDiff     `json:"schema,omitempty" yaml:"schema,omitempty"`
	ContentDiff     *ContentDiff    `json:"content,omitempty" yaml:"content,omitempty"`
}

// Empty indicates whether a change was found in this element
func (headerDiff *HeaderDiff) Empty() bool {
	return headerDiff == nil || *headerDiff == HeaderDiff{}
}

func getHeaderDiff(config *Config, state *state, header1, header2 *openapi3.Header) (*HeaderDiff, error) {
	diff, err := getHeaderDiffInternal(config, state, header1, header2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getHeaderDiffInternal(config *Config, state *state, header1, header2 *openapi3.Header) (*HeaderDiff, error) {
	result := HeaderDiff{}
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, header1.Extensions, header2.Extensions)
	if err != nil {
		return nil, err
	}
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), header1.Description, header2.Description)
	result.DeprecatedDiff = getValueDiff(header1.Deprecated, header2.Deprecated)
	result.RequiredDiff = getValueDiff(header1.Required, header2.Required)

	result.ExamplesDiff, err = getExamplesDiff(config, header1.Examples, header2.Examples)
	if err != nil {
		return nil, err
	}

	result.SchemaDiff, err = getSchemaDiff(config, state, header1.Schema, header2.Schema)
	if err != nil {
		return nil, err
	}

	result.ExampleDiff = getValueDiffConditional(config.IsExcludeExamples(), header1.Example, header2.Example)

	result.ContentDiff, err = getContentDiff(config, state, header1.Content, header2.Content)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
