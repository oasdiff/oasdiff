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

// walkModifiedParameters invokes processor for every modified parameter of
// every modified operation.
func walkModifiedParameters(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config, processor func(p paramInfo)) {
	if diffReport.PathsDiff == nil {
		return
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil || operationItem.ParametersDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for location, params := range operationItem.ParametersDiff.Modified {
				for name, paramDiff := range params {
					processor(paramInfo{
						opInfo:    opInfo,
						location:  location,
						name:      name,
						paramDiff: paramDiff,
					})
				}
			}
		}
	}
}

// headerInfo is what walkModifiedResponseHeaders hands its processor.
type headerInfo struct {
	opInfo         opInfo
	responseStatus string
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
						name:           name,
						headerDiff:     headerDiff,
					})
				}
			}
		}
	}
}
