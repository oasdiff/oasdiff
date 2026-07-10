package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxLengthDecreasedId = "request-parameter-max-length-decreased"
	RequestParameterMaxLengthIncreasedId = "request-parameter-max-length-increased"
)

func RequestParameterMaxLengthUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {
				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}
					maxLengthDiff := paramDiff.SchemaDiff.MaxLengthDiff
					if maxLengthDiff == nil {
						continue
					}
					if maxLengthDiff.From == nil ||
						maxLengthDiff.To == nil {
						continue
					}

					id := RequestParameterMaxLengthDecreasedId
					if !isDecreasedValue(maxLengthDiff) {
						id = RequestParameterMaxLengthIncreasedId
					}

					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "maxLength")
					result = append(result, opInfo.NewApiChange(
						id,
						[]any{paramLocation, paramName, maxLengthDiff.From, maxLengthDiff.To},
						"",
					).WithSources(baseSource, revisionSource))
				}
			}
		}
	}
	return result
}
