package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxLengthSetId = "request-parameter-max-length-set"
)

func RequestParameterMaxLengthSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					maxLengthDiff := paramDiff.SchemaDiff.MaxLengthDiff
					if maxLengthDiff == nil {
						continue
					}
					if maxLengthDiff.From != nil ||
						maxLengthDiff.To == nil {
						continue
					}

					_, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "maxLength")
					result = append(result, NewApiChange(
						RequestParameterMaxLengthSetId,
						config,
						[]any{paramLocation, paramName, maxLengthDiff.To},
						commentId(RequestParameterMaxLengthSetId),
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithSources(nil, revisionSource))
				}
			}
		}
	}
	return result
}
