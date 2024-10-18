package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// TagDiff describes the changes between a pair of tag objects: https://swagger.io/specification/#tag-object
type TagDiff struct {
	NameDiff        *ValueDiff `json:"name,omitempty" yaml:"name,omitempty"`
	DescriptionDiff *ValueDiff `json:"description,omitempty" yaml:"description,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *TagDiff) Empty() bool {
	return diff == nil || *diff == TagDiff{}
}

func getTagDiff(config *Config, tag1, tag2 *openapi3.Tag) *TagDiff {
	diff := getTagDiffInternal(config, tag1, tag2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getTagDiffInternal(config *Config, tag1, tag2 *openapi3.Tag) *TagDiff {
	result := TagDiff{}

	result.NameDiff = getValueDiff(tag1.Name, tag2.Name)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), tag1.Description, tag2.Description)

	return &result
}
