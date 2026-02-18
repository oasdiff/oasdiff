package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxItemsIncreasedId = "request-parameter-max-items-increased"
	RequestParameterMaxItemsDecreasedId = "request-parameter-max-items-decreased"
)

func RequestParameterMaxItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if paramDiff.SchemaDiff == nil || paramDiff.SchemaDiff.ItemsDiff == nil {
						continue
					}
					maxItemsDiff := paramDiff.SchemaDiff.ItemsDiff.MaxItemsDiff
					if maxItemsDiff == nil {
						continue
					}
					if maxItemsDiff.From == nil ||
						maxItemsDiff.To == nil {
						continue
					}

					id := RequestParameterMaxItemsDecreasedId
					if IsIncreasedValue(maxItemsDiff) {
						id = RequestParameterMaxItemsIncreasedId
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{paramLocation, paramName, maxItemsDiff.From, maxItemsDiff.To},
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
