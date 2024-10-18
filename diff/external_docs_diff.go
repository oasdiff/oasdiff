package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ExternalDocsDiff describes the changes between a pair of external documentation objects: https://swagger.io/specification/#external-documentation-object
type ExternalDocsDiff struct {
	Added           bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted         bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff  *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	DescriptionDiff *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	URLDiff         *ValueDiff      `json:"url,omitempty" yaml:"url,omitempty"`
}

func newExternalDocsDiff() *ExternalDocsDiff {
	return &ExternalDocsDiff{}
}

// Empty indicates whether a change was found in this element
func (diff *ExternalDocsDiff) Empty() bool {
	return diff == nil || *diff == ExternalDocsDiff{}
}

func getExternalDocsDiff(config *Config, docs1, docs2 *openapi3.ExternalDocs) (*ExternalDocsDiff, error) {
	diff, err := getExternalDocsDiffInternal(config, docs1, docs2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getExternalDocsDiffInternal(config *Config, docs1, docs2 *openapi3.ExternalDocs) (*ExternalDocsDiff, error) {
	result := newExternalDocsDiff()
	var err error

	if docs1 == nil && docs2 == nil {
		return result, nil
	}

	if docs1 == nil && docs2 != nil {
		result.Added = true
		return result, nil
	}

	if docs1 != nil && docs2 == nil {
		result.Deleted = true
		return result, nil
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, docs1.Extensions, docs2.Extensions)
	if err != nil {
		return nil, err
	}
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), docs1.Description, docs2.Description)
	result.URLDiff = getValueDiff(docs1.URL, docs2.URL)

	return result, nil
}
