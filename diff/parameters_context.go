package diff

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/commonparams"
)

// ParametersContext holds the parameters of each operation as a request
// carries them, for the checks. It differs from the parameters in PathsDiff,
// which are compared as written, in two ways:
//   - the parameters of a path item are merged into each of its operations, so
//     moving a parameter between a path item and its operations is no change
//   - header parameters are matched by name ignoring case, so renaming a header
//     in case only is no change
type ParametersContext struct {
	// Operations holds the operations present in both specs whose parameters
	// changed: those of paths, then those of webhooks.
	Operations []OperationParameters
}

// OperationParameters holds the parameters of one operation in both specs, and
// their diff. The parameters are the operation's own and those of its path
// item that it doesn't override. Header parameters are matched by name
// ignoring case, since HTTP header names are case-insensitive.
type OperationParameters struct {
	OperationPair
	// Webhook is true for an operation of a webhook. Path is then the
	// webhook's name.
	Webhook            bool
	BaseParameters     openapi3.Parameters
	RevisionParameters openapi3.Parameters
	ParametersDiff     *ParametersDiffByLocation
}

// Without returns a copy of the context that leaves out the operations remove
// reports. The receiver is not modified.
func (parametersContext *ParametersContext) Without(remove func(OperationPair) bool) *ParametersContext {
	if parametersContext == nil {
		return nil
	}
	result := *parametersContext
	result.Operations = slices.DeleteFunc(slices.Clone(parametersContext.Operations), func(operation OperationParameters) bool {
		return remove(operation.OperationPair)
	})
	return &result
}

func getParametersContext(config *Config, state *state, paths, webhooks pathItemPairs) (*ParametersContext, error) {
	operations, err := getOperationsParameters(config, state, paths, false)
	if err != nil {
		return nil, err
	}
	webhookOperations, err := getOperationsParameters(config, state, webhooks, true)
	if err != nil {
		return nil, err
	}
	return &ParametersContext{Operations: append(operations, webhookOperations...)}, nil
}

func getOperationsParameters(config *Config, state *state, pathItemPairs pathItemPairs, webhook bool) ([]OperationParameters, error) {
	result := []OperationParameters{}
	for _, pair := range getOperationPairs(pathItemPairs) {
		pathItemPair := pathItemPairs[pair.Path]
		params1 := commonparams.Merge(pathItemPair.PathItem1.Parameters, pair.Base.Parameters)
		params2 := commonparams.Merge(pathItemPair.PathItem2.Parameters, pair.Revision.Parameters)
		paramsDiff, err := getParametersDiffByLocation(config, state, params1, params2, pathItemPair.PathParamsMap, headerNamesIgnoringCase)
		if err != nil {
			return nil, err
		}
		if paramsDiff.Empty() {
			continue
		}
		result = append(result, OperationParameters{
			OperationPair:      pair,
			Webhook:            webhook,
			BaseParameters:     params1,
			RevisionParameters: params2,
			ParametersDiff:     paramsDiff,
		})
	}
	return result, nil
}
