package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// MethodDiff describes the changes between a pair of operation objects: https://swagger.io/specification/#operation-object
type MethodDiff struct {
	ExtensionsDiff   *ExtensionsDiff           `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	TagsDiff         *StringsDiff              `json:"tags,omitempty" yaml:"tags,omitempty"`
	SummaryDiff      *ValueDiff                `json:"summary,omitempty" yaml:"summary,omitempty"`
	DescriptionDiff  *ValueDiff                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationIDDiff  *ValueDiff                `json:"operationID,omitempty" yaml:"operationID,omitempty"`
	ParametersDiff   *ParametersDiffByLocation `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBodyDiff  *RequestBodyDiff          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	ResponsesDiff    *ResponsesDiff            `json:"responses,omitempty" yaml:"responses,omitempty"`
	CallbacksDiff    *CallbacksDiff            `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	DeprecatedDiff   *ValueDiff                `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	SecurityDiff     *SecurityRequirementsDiff `json:"securityRequirements,omitempty" yaml:"securityRequirements,omitempty"`
	ServersDiff      *ServersDiff              `json:"servers,omitempty" yaml:"servers,omitempty"`
	ExternalDocsDiff *ExternalDocsDiff         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Base             *openapi3.Operation       `json:"-" yaml:"-"`
	Revision         *openapi3.Operation       `json:"-" yaml:"-"`
}

func newMethodDiff() *MethodDiff {
	return &MethodDiff{}
}

// Empty indicates whether a change was found in this element
func (methodDiff *MethodDiff) Empty() bool {
	if methodDiff == nil {
		return true
	}

	return *methodDiff == MethodDiff{Base: methodDiff.Base, Revision: methodDiff.Revision}
}

func getMethodDiff(config *Config, state *state, operation1, operation2 *openapi3.Operation, pathParamsMap PathParamsMap) (*MethodDiff, error) {

	diff, err := getMethodDiffInternal(config, state, operation1, operation2, pathParamsMap)

	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getMethodDiffInternal(config *Config, state *state, operation1, operation2 *openapi3.Operation, pathParamsMap PathParamsMap) (*MethodDiff, error) {

	result := newMethodDiff()
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, operation1.Extensions, operation2.Extensions)
	if err != nil {
		return nil, err
	}
	result.TagsDiff = getStringsDiff(operation1.Tags, operation2.Tags)
	result.SummaryDiff = getValueDiffConditional(config.IsExcludeSummary(), operation1.Summary, operation2.Summary)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), operation1.Description, operation2.Description)
	result.OperationIDDiff = getValueDiff(operation1.OperationID, operation2.OperationID)
	result.ParametersDiff, err = getParametersDiffByLocation(config, state, operation1.Parameters, operation2.Parameters, pathParamsMap)
	if err != nil {
		return nil, err
	}

	result.RequestBodyDiff, err = getRequestBodyDiff(config, state, operation1.RequestBody, operation2.RequestBody)
	if err != nil {
		return nil, err
	}

	result.ResponsesDiff, err = getResponsesDiff(config, state, operation1.Responses, operation2.Responses)
	if err != nil {
		return nil, err
	}

	result.CallbacksDiff, err = getCallbacksDiff(config, state, operation1.Callbacks, operation2.Callbacks)
	if err != nil {
		return nil, err
	}
	result.DeprecatedDiff = getValueDiff(operation1.Deprecated, operation2.Deprecated)
	result.SecurityDiff = getSecurityRequirementsDiff(operation1.Security, operation2.Security)
	result.ServersDiff = getServersDiff(config, operation1.Servers, operation2.Servers)
	result.ExternalDocsDiff, err = getExternalDocsDiff(config, operation1.ExternalDocs, operation2.ExternalDocs)
	if err != nil {
		return nil, err
	}
	result.Base = operation1
	result.Revision = operation2

	return result, nil
}

// Patch applies the patch to a method
func (methodDiff *MethodDiff) Patch(operation *openapi3.Operation) error {

	if methodDiff.Empty() {
		return nil
	}

	if err := methodDiff.DescriptionDiff.patchString(&operation.Description); err != nil {
		return err
	}

	if err := methodDiff.ParametersDiff.Patch(operation.Parameters); err != nil {
		return err
	}

	return nil
}
