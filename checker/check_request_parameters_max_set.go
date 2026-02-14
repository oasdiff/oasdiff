package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxSetId          = "request-parameter-max-set"
	RequestParameterExclusiveMaxSetId = "request-parameter-exclusive-max-set"
)

func RequestParameterMaxSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
						maxDiff.From == nil &&
						maxDiff.To != nil {

						result = append(result, NewApiChange(
							RequestParameterMaxSetId,
							config,
							[]any{paramLocation, paramName, maxDiff.To},
							commentId(RequestParameterMaxSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					exMaxDiff := paramDiff.SchemaDiff.ExclusiveMaxDiff
					if exMaxDiff != nil &&
						exMaxDiff.From == nil &&
						exMaxDiff.To != nil {

						result = append(result, NewApiChange(
							RequestParameterExclusiveMaxSetId,
							config,
							[]any{paramLocation, paramName, exMaxDiff.To},
							commentId(RequestParameterExclusiveMaxSetId),
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
