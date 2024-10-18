package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ServerDiff describes the changes between a pair of server objects: https://swagger.io/specification/#server-object
type ServerDiff struct {
	Added           bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted         bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff  *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	URLDiff         *ValueDiff      `json:"urlType,omitempty" yaml:"urlType,omitempty"`
	DescriptionDiff *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	VariablesDiff   *VariablesDiff  `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *ServerDiff) Empty() bool {
	return diff == nil || *diff == ServerDiff{}
}

func getServerDiff(config *Config, value1, value2 *openapi3.Server) (*ServerDiff, error) {
	diff, err := getServerDiffInternal(config, value1, value2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getServerDiffInternal(config *Config, value1, value2 *openapi3.Server) (*ServerDiff, error) {

	if value1 == nil && value2 == nil {
		return nil, nil
	}

	result := ServerDiff{}
	var err error

	if value1 == nil && value2 != nil {
		result.Added = true
		return &result, nil
	}

	if value1 != nil && value2 == nil {
		result.Deleted = true
		return &result, nil
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, value1.Extensions, value2.Extensions)
	if err != nil {
		return nil, err
	}

	result.URLDiff = getValueDiff(value1.URL, value2.URL)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), value1.Description, value2.Description)
	result.VariablesDiff, err = getVariablesDiff(config, value1.Variables, value2.Variables)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
