package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMinItemsSetId = "request-parameter-min-items-set"
)

func RequestParameterMinItemsSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					_, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "minItems")
					minItemsDiff := paramDiff.SchemaDiff.MinItemsDiff
					if minItemsDiff == nil {
						continue
					}
					if minItemsDiff.From != nil ||
						minItemsDiff.To == nil {
						continue
					}

					result = append(result, NewApiChange(
						RequestParameterMinItemsSetId,
						config,
						[]any{paramLocation, paramName, minItemsDiff.To},
						commentId(RequestParameterMinItemsSetId),
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
