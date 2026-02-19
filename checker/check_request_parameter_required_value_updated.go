package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecomeRequiredId = "request-parameter-became-required"
	RequestParameterBecomeOptionalId = "request-parameter-became-optional"
)

func RequestParameterRequiredValueUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil {
				continue
			}
			if operationItem.ParametersDiff.Modified == nil {
				continue
			}
			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for paramLocation, paramItems := range operationItem.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {
					requiredDiff := paramItem.RequiredDiff
					if requiredDiff == nil {
						continue
					}

					id := RequestParameterBecomeRequiredId

					if requiredDiff.To != true {
						id = RequestParameterBecomeOptionalId
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{paramLocation, paramName},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithSources(baseSource, revisionSource))
				}
			}
		}
	}
	return result
}
