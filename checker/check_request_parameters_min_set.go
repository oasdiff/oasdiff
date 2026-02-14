package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinSetId          = "request-parameter-min-set"
	RequestParameterExclusiveMinSetId = "request-parameter-exclusive-min-set"
)

func RequestParameterMinSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
						minDiff.From == nil &&
						minDiff.To != nil {

						result = append(result, NewApiChange(
							RequestParameterMinSetId,
							config,
							[]any{paramLocation, paramName, minDiff.To},
							commentId(RequestParameterMinSetId),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}

					exMinDiff := paramDiff.SchemaDiff.ExclusiveMinDiff
					if exMinDiff != nil &&
						exMinDiff.From == nil &&
						exMinDiff.To != nil {

						result = append(result, NewApiChange(
							RequestParameterExclusiveMinSetId,
							config,
							[]any{paramLocation, paramName, exMinDiff.To},
							commentId(RequestParameterExclusiveMinSetId),
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
