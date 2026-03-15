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
					for _, entry := range []struct {
						diff  *diff.ValueDiff
						id    string
						field string
					}{
						{paramDiff.SchemaDiff.MinDiff, RequestParameterMinSetId, "minimum"},
						{paramDiff.SchemaDiff.ExclusiveMinDiff, RequestParameterExclusiveMinSetId, "exclusiveMinimum"},
					} {
						if entry.diff == nil || entry.diff.From != nil || entry.diff.To == nil {
							continue
						}
						_, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, entry.field)
						result = append(result, NewApiChange(
							entry.id,
							config,
							[]any{paramLocation, paramName, entry.diff.To},
							commentId(entry.id),
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(nil, revisionSource))
					}
				}
			}
		}
	}
	return result
}
