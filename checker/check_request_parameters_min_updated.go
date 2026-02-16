package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinIncreasedId          = "request-parameter-min-increased"
	RequestParameterMinDecreasedId          = "request-parameter-min-decreased"
	RequestParameterExclusiveMinIncreasedId = "request-parameter-exclusive-min-increased"
	RequestParameterExclusiveMinDecreasedId = "request-parameter-exclusive-min-decreased"
)

func RequestParameterMinUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {
				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}
					minDiff := paramDiff.SchemaDiff.MinDiff
					if minDiff != nil &&
						minDiff.From != nil &&
						minDiff.To != nil {

						id := RequestParameterMinIncreasedId
						if !IsIncreasedValue(minDiff) {
							id = RequestParameterMinDecreasedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, minDiff.From, minDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					exMinDiff := paramDiff.SchemaDiff.ExclusiveMinDiff
					if exMinDiff != nil &&
						exMinDiff.From != nil &&
						exMinDiff.To != nil {

						id := RequestParameterExclusiveMinIncreasedId
						if !IsIncreasedValue(exMinDiff) {
							id = RequestParameterExclusiveMinDecreasedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, exMinDiff.From, exMinDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}
				}
			}
		}
	}
	return result
}
