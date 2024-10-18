package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// MediaTypeDiff describes the changes between a pair of media type objects
type MediaTypeDiff struct {
	ExtensionsDiff *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	SchemaDiff     *SchemaDiff     `json:"schema,omitempty" yaml:"schema,omitempty"`
	ExampleDiff    *ValueDiff      `json:"example,omitempty" yaml:"example,omitempty"`
	ExamplesDiff   *ExamplesDiff   `json:"examples,omitempty" yaml:"examples,omitempty"`
	EncodingsDiff  *EncodingsDiff  `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *MediaTypeDiff) Empty() bool {
	return diff == nil || *diff == MediaTypeDiff{}
}

func getMediaTypeDiff(config *Config, state *state, mediaType1 *openapi3.MediaType, mediaType2 *openapi3.MediaType) (*MediaTypeDiff, error) {
	diff, err := getMediaTypeDiffInternal(config, state, mediaType1, mediaType2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getMediaTypeDiffInternal(config *Config, state *state, mediaType1 *openapi3.MediaType, mediaType2 *openapi3.MediaType) (*MediaTypeDiff, error) {
	result := MediaTypeDiff{}
	var err error

	if mediaType1 == nil || mediaType2 == nil {
		return nil, fmt.Errorf("media type is nil")
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, mediaType1.Extensions, mediaType2.Extensions)
	if err != nil {
		return nil, err
	}
	result.SchemaDiff, err = getSchemaDiff(config, state, mediaType1.Schema, mediaType2.Schema)
	if err != nil {
		return nil, err
	}
	result.ExampleDiff = getValueDiffConditional(config.IsExcludeExamples(), mediaType1.Example, mediaType2.Example)
	result.EncodingsDiff, err = getEncodingsDiff(config, state, mediaType1.Encoding, mediaType2.Encoding)
	if err != nil {
		return nil, err
	}
	result.ExamplesDiff, err = getExamplesDiff(config, mediaType1.Examples, mediaType2.Examples)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
