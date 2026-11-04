package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/commonparams"
)

// ParametersContext carries the parameters each operation has, for checkers.
// PathsDiff compares parameters where they are written, so a parameter moved
// between a path item and its operations is a change there and none here.
type ParametersContext struct {
	// Operations and Webhooks hold the operations present in both specs whose
	// parameters changed, sorted by path, or webhook name, and method.
	Operations []OperationParameters
	Webhooks   []OperationParameters
}

// OperationParameters compares the parameters an operation has: its own, and
// those of its path item that it doesn't override. A request to the operation
// carries both, so a change to a path-level parameter, or moving a parameter
// between the path item and the operation, shows here as it reaches the
// operation.
type OperationParameters struct {
	OperationPair
	BaseParameters     openapi3.Parameters
	RevisionParameters openapi3.Parameters
	ParametersDiff     *ParametersDiffByLocation
}

func getParametersContext(config *Config, state *state, paths, webhooks pathItemPairs) (*ParametersContext, error) {
	operations, err := getOperationsParameters(config, state, paths)
	if err != nil {
		return nil, err
	}
	webhookOperations, err := getOperationsParameters(config, state, webhooks)
	if err != nil {
		return nil, err
	}
	return &ParametersContext{Operations: operations, Webhooks: webhookOperations}, nil
}

func getOperationsParameters(config *Config, state *state, pathItemPairs pathItemPairs) ([]OperationParameters, error) {
	result := []OperationParameters{}
	for _, pair := range getOperationPairs(pathItemPairs) {
		pathItemPair := pathItemPairs[pair.Path]
		params1 := commonparams.Merge(pathItemPair.PathItem1.Parameters, pair.Base.Parameters)
		params2 := commonparams.Merge(pathItemPair.PathItem2.Parameters, pair.Revision.Parameters)
		paramsDiff, err := getParametersDiffByLocation(config, state, params1, params2, pathItemPair.PathParamsMap)
		if err != nil {
			return nil, err
		}
		if paramsDiff.Empty() {
			continue
		}
		result = append(result, OperationParameters{
			OperationPair:      pair,
			BaseParameters:     params1,
			RevisionParameters: params2,
			ParametersDiff:     paramsDiff,
		})
	}
	return result, nil
}

// getWebhookPairs matches webhooks by name, as WebhooksDiff does.
func getWebhookPairs(webhooks1, webhooks2 map[string]*openapi3.PathItem) pathItemPairs {
	result := pathItemPairs{}
	_, _, common := keysDiff(webhooks1, webhooks2)
	for _, name := range common {
		result[name] = &pathItemPair{
			PathItem1:     webhooks1[name],
			PathItem2:     webhooks2[name],
			PathParamsMap: PathParamsMap{},
		}
	}
	return result
}
