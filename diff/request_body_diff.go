package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// RequestBodyDiff describes the changes between a pair of request body objects: https://swagger.io/specification/#request-body-object
type RequestBodyDiff struct {
	Added           bool            `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted         bool            `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ExtensionsDiff  *ExtensionsDiff `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	DescriptionDiff *ValueDiff      `json:"description,omitempty" yaml:"description,omitempty"`
	RequiredDiff    *ValueDiff      `json:"required,omitempty" yaml:"required,omitempty"`
	ContentDiff     *ContentDiff    `json:"content,omitempty" yaml:"content,omitempty"`
}

// Empty indicates whether a change was found in this element
func (diff *RequestBodyDiff) Empty() bool {
	if diff == nil {
		return true
	}
	return *diff == RequestBodyDiff{}
}

func newRequestBodyDiff() *RequestBodyDiff {
	return &RequestBodyDiff{}
}

func getRequestBodyDiff(config *Config, state *state, requestBodyRef1, requestBodyRef2 *openapi3.RequestBodyRef) (*RequestBodyDiff, error) {
	diff, err := getRequestBodyDiffInternal(config, state, requestBodyRef1, requestBodyRef2)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getRequestBodyDiffInternal(config *Config, state *state, requestBodyRef1, requestBodyRef2 *openapi3.RequestBodyRef) (*RequestBodyDiff, error) {

	if requestBodyRef1 == nil && requestBodyRef2 == nil {
		return nil, nil
	}

	result := newRequestBodyDiff()

	if requestBodyRef1 == nil && requestBodyRef2 != nil {
		result.Added = true
		return result, nil
	}

	if requestBodyRef1 != nil && requestBodyRef2 == nil {
		result.Deleted = true
		return result, nil
	}

	requestBody1, err := derefRequestBody(requestBodyRef1)
	if err != nil {
		return nil, err
	}
	requestBody2, err := derefRequestBody(requestBodyRef2)
	if err != nil {
		return nil, err
	}

	result.ExtensionsDiff, err = getExtensionsDiff(config, requestBody1.Extensions, requestBody2.Extensions)
	if err != nil {
		return nil, err
	}

	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), requestBody1.Description, requestBody2.Description)
	result.RequiredDiff = getValueDiff(requestBody1.Required, requestBody2.Required)
	result.ContentDiff, err = getContentDiff(config, state, requestBody1.Content, requestBody2.Content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func derefRequestBody(ref *openapi3.RequestBodyRef) (*openapi3.RequestBody, error) {

	if ref == nil || ref.Value == nil {
		return nil, fmt.Errorf("request body reference is nil")
	}

	return ref.Value, nil
}
