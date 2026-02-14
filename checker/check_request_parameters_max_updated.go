package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxDecreasedId          = "request-parameter-max-decreased"
	RequestParameterMaxIncreasedId          = "request-parameter-max-increased"
	RequestParameterExclusiveMaxDecreasedId = "request-parameter-exclusive-max-decreased"
	RequestParameterExclusiveMaxIncreasedId = "request-parameter-exclusive-max-increased"
)

func RequestParameterMaxUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					maxDiff := paramDiff.SchemaDiff.MaxDiff
					if maxDiff != nil &&
						maxDiff.From != nil &&
						maxDiff.To != nil {

						id := RequestParameterMaxDecreasedId

						if !IsDecreasedValue(maxDiff) {
							id = RequestParameterMaxIncreasedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, maxDiff.From, maxDiff.To},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					exMaxDiff := paramDiff.SchemaDiff.ExclusiveMaxDiff
					if exMaxDiff != nil &&
						exMaxDiff.From != nil &&
						exMaxDiff.To != nil {

						id := RequestParameterExclusiveMaxDecreasedId

						if !IsDecreasedValue(exMaxDiff) {
							id = RequestParameterExclusiveMaxIncreasedId
						}

						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, exMaxDiff.From, exMaxDiff.To},
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
