package diff

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// PathDiff describes the changes between a pair of path item objects: https://swagger.io/specification/#path-item-object
type PathDiff struct {
	ExtensionsDiff  *ExtensionsDiff           `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	RefDiff         *ValueDiff                `json:"ref,omitempty" yaml:"ref,omitempty"`
	SummaryDiff     *ValueDiff                `json:"summary,omitempty" yaml:"summary,omitempty"`
	DescriptionDiff *ValueDiff                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationsDiff  *OperationsDiff           `json:"operations,omitempty" yaml:"operations,omitempty"`
	ServersDiff     *ServersDiff              `json:"servers,omitempty" yaml:"servers,omitempty"`
	ParametersDiff  *ParametersDiffByLocation `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Base            *openapi3.PathItem        `json:"-" yaml:"-"`
	Revision        *openapi3.PathItem        `json:"-" yaml:"-"`
}

func newPathDiff() *PathDiff {
	return &PathDiff{}
}

// Empty indicates whether a change was found in this element
func (pathDiff *PathDiff) Empty() bool {
	return pathDiff == nil || *pathDiff == PathDiff{Base: pathDiff.Base, Revision: pathDiff.Revision}
}

func getPathDiff(config *Config, state *state, pathItemPair *pathItemPair) (*PathDiff, error) {

	diff, err := getPathDiffInternal(config, state, pathItemPair)
	if err != nil {
		return nil, err
	}

	if diff.Empty() {
		return nil, nil
	}

	return diff, nil
}

func getPathDiffInternal(config *Config, state *state, pathItemPair *pathItemPair) (*PathDiff, error) {

	pathItem1 := pathItemPair.PathItem1
	pathItem2 := pathItemPair.PathItem2

	if pathItem1 == nil || pathItem2 == nil {
		return nil, fmt.Errorf("path item is nil")
	}

	result := newPathDiff()
	var err error

	result.ExtensionsDiff, err = getExtensionsDiff(config, pathItem1.Extensions, pathItem2.Extensions)
	if err != nil {
		return nil, err
	}

	result.RefDiff = getValueDiff(pathItem1.Ref, pathItem2.Ref)
	result.SummaryDiff = getValueDiffConditional(config.IsExcludeSummary(), pathItem1.Summary, pathItem2.Summary)
	result.DescriptionDiff = getValueDiffConditional(config.IsExcludeDescription(), pathItem1.Description, pathItem2.Description)

	result.OperationsDiff, err = getOperationsDiff(config, state, pathItemPair)
	if err != nil {
		return nil, err
	}

	result.ServersDiff = getServersDiff(config, &pathItem1.Servers, &pathItem2.Servers)
	result.ParametersDiff, err = getParametersDiffByLocation(config, state, pathItem1.Parameters, pathItem2.Parameters, pathItemPair.PathParamsMap)
	result.Base = pathItem1
	result.Revision = pathItem2
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Patch applies the patch to a path item
func (pathDiff *PathDiff) Patch(pathItem *openapi3.PathItem) error {

	if pathDiff.Empty() {
		return nil
	}

	err := pathDiff.OperationsDiff.Patch(pathItem.Operations())
	if err != nil {
		return err
	}

	return err
}
