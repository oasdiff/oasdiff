package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

// paramInfo is what walkModifiedParameters hands its processor.
type paramInfo struct {
	opInfo    opInfo
	location  string
	name      string
	paramDiff *diff.ParameterDiff
}

// operationParams is what walkOperationsParameters hands its processor.
type operationParams struct {
	opInfo opInfo
	diff.OperationParameters
}

// walkOperationsParameters invokes processor for every operation whose
// parameters changed, comparing the parameters each operation has, its own and
// its path item's, rather than each list where it is written.
func walkOperationsParameters(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config, processor func(o operationParams)) {
	parametersContext := diffReport.ParametersContext
	if parametersContext == nil {
		return
	}
	for _, operation := range parametersContext.Operations {
		path := operation.Path
		if operation.Webhook {
			path = webhookPathPrefix + path
		}
		processor(operationParams{
			opInfo:              newOpInfoFromDiff(config, operation.OperationPair, operationsSources, operation.Method, path),
			OperationParameters: operation,
		})
	}
}

// walkModifiedParameters invokes processor for every modified parameter of
// every operation whose parameters changed.
func walkModifiedParameters(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config, processor func(p paramInfo)) {
	walkOperationsParameters(diffReport, operationsSources, config, func(o operationParams) {
		for location, params := range o.ParametersDiff.Modified {
			for name, paramDiff := range params {
				processor(paramInfo{
					opInfo:    o.opInfo,
					location:  location,
					name:      name,
					paramDiff: paramDiff,
				})
			}
		}
	})
}
