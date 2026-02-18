package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinItemsIncreasedId = "request-parameter-min-items-increased"
	RequestParameterMinItemsDecreasedId = "request-parameter-min-items-decreased"
)

func RequestParameterMinItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {
				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}
					minItemsDiff := paramDiff.SchemaDiff.MinItemsDiff
					if minItemsDiff == nil {
						continue
					}
					if minItemsDiff.From == nil ||
						minItemsDiff.To == nil {
						continue
					}

					id := RequestParameterMinItemsIncreasedId
					if !IsIncreasedValue(minItemsDiff) {
						id = RequestParameterMinItemsDecreasedId
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{paramLocation, paramName, minItemsDiff.From, minItemsDiff.To},
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
