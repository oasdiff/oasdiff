package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ParameterDiff describes the changes between a pair of parameter objects: https://swagger.io/specification/#parameter-object
type ParameterDiff struct {
	NameDiff            *ValueDiff          `json:"name,omitempty" yaml:"name,omitempty"`
	InDiff              *ValueDiff          `json:"in,omitempty" yaml:"in,omitempty"`
	ExtensionsDiff      *ExtensionsDiff     `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	DescriptionDiff     *ValueDiff          `json:"description,omitempty" yaml:"description,omitempty"`
	StyleDiff           *ValueDiff          `json:"style,omitempty" yaml:"style,omitempty"`
	ExplodeDiff         *ValueDiff          `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowEmptyValueDiff *ValueDiff          `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	AllowReservedDiff   *ValueDiff          `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	DeprecatedDiff      *ValueDiff          `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	RequiredDiff        *ValueDiff          `json:"required,omitempty" yaml:"required,omitempty"`
	SchemaDiff          *SchemaDiff         `json:"schema,omitempty" yaml:"schema,omitempty"`
	ExampleDiff         *ValueDiff          `json:"example,omitempty" yaml:"example,omitempty"`
	ExamplesDiff        *ExamplesDiff       `json:"examples,omitempty" yaml:"examples,omitempty"`
	ContentDiff         *ContentDiff        `json:"content,omitempty" yaml:"content,omitempty"`
	Base                *openapi3.Parameter `json:"-" yaml:"-"`
	Revision            *openapi3.Parameter `json:"-" yaml:"-"`
}

// Empty indicates whether a change was found in this element
func (diff *ParameterDiff) Empty() bool {
	return diff == nil || *diff == ParameterDiff{Base: diff.Base, Revision: diff.Revision}
}

func getParameterDiff(config *Config, state *state, param1, param2 *openapi3.Parameter) (*ParameterDiff, error) {
	diff, err := getParameterDiffInternal(config, state, param1, param2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getParameterDiffInternal(config *Config, state *state, param1, param2 *openapi3.Parameter) (*ParameterDiff, error) {

	result := ParameterDiff{}
	var err error

	result.NameDiff = getValueDiff(param1.Name, param2.Name)
	result.InDiff = getValueDiff(param1.In, param2.In)
	result.ExtensionsDiff, err = getExtensionsDiff(config, param1.Extensions, param2.Extensions)
	if err != nil {
		return nil, err
	}

	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), param1.Description, param2.Description)
	result.StyleDiff = getValueDiff(param1.Style, param2.Style)
	result.ExplodeDiff = getBoolRefDiff(param1.Explode, param2.Explode)
	result.AllowEmptyValueDiff = getValueDiff(param1.AllowEmptyValue, param2.AllowEmptyValue)
	result.AllowReservedDiff = getValueDiff(param1.AllowReserved, param2.AllowReserved)
	result.DeprecatedDiff = getValueDiff(param1.Deprecated, param2.Deprecated)
	result.RequiredDiff = getValueDiff(param1.Required, param2.Required)
	result.SchemaDiff, err = getSchemaDiff(config, state, param1.Schema, param2.Schema)
	if err != nil {
		return nil, err
	}

	result.ExampleDiff = getValueDiffConditional(config.IsExcludeExamples(), param1.Example, param2.Example)

	result.ExamplesDiff, err = getExamplesDiff(config, param1.Examples, param2.Examples)
	if err != nil {
		return nil, err
	}

	result.ContentDiff, err = getContentDiff(config, state, param1.Content, param2.Content)
	if err != nil {
		return nil, err
	}
	result.Base = param1
	result.Revision = param2

	return &result, nil
}

// Patch applies the patch to a parameter
func (diff *ParameterDiff) Patch(parameter *openapi3.Parameter) error {

	if err := diff.DescriptionDiff.patchString(&parameter.Description); err != nil {
		return err
	}

	schema, err := derefSchema(parameter.Schema)
	if err != nil {
		// no schema to patch, continue.
		return nil
	}

	return diff.SchemaDiff.Patch(schema)
}
