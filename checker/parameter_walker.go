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
func walkOperationsParameters(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config, processor func(o operationParams)) {
	parametersContext := diffReport.ParametersContext
	if parametersContext == nil {
		return
	}
	walk := func(pathPrefix string, operations []diff.OperationParameters) {
		for _, operation := range operations {
			// the source lookups read only Base and Revision
			methodDiff := &diff.MethodDiff{Base: operation.Base, Revision: operation.Revision}
			processor(operationParams{
				opInfo:              newOpInfoFromDiff(config, methodDiff, operationsSources, operation.Method, pathPrefix+operation.Path),
				OperationParameters: operation,
			})
		}
	}
	walk("", parametersContext.Operations)
	walk(webhookPathPrefix, parametersContext.Webhooks)
}

// walkModifiedParameters invokes processor for every modified parameter of
// every operation whose parameters changed.
func walkModifiedParameters(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config, processor func(p paramInfo)) {
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

// headerInfo is what walkModifiedResponseHeaders hands its processor.
type headerInfo struct {
	opInfo         opInfo
	responseStatus string
	responseDiff   *diff.ResponseDiff
	name           string
	headerDiff     *diff.HeaderDiff
}

// walkModifiedResponseHeaders invokes processor for every modified header of
// every modified response of every modified operation.
func walkModifiedResponseHeaders(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config, processor func(h headerInfo)) {
	if diffReport.PathsDiff == nil {
		return
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.HeadersDiff == nil {
					continue
				}
				for name, headerDiff := range responseDiff.HeadersDiff.Modified {
					processor(headerInfo{
						opInfo:         opInfo,
						responseStatus: responseStatus,
						responseDiff:   responseDiff,
						name:           name,
						headerDiff:     headerDiff,
					})
				}
			}
		}
	}
}
