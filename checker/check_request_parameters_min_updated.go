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
					for _, entry := range []struct {
						diff        *diff.ValueDiff
						increasedId string
						decreasedId string
						field       string
					}{
						{paramDiff.SchemaDiff.MinDiff, RequestParameterMinIncreasedId, RequestParameterMinDecreasedId, "minimum"},
						{paramDiff.SchemaDiff.ExclusiveMinDiff, RequestParameterExclusiveMinIncreasedId, RequestParameterExclusiveMinDecreasedId, "exclusiveMinimum"},
					} {
						if entry.diff == nil || entry.diff.From == nil || entry.diff.To == nil {
							continue
						}
						id := entry.increasedId
						if !IsIncreasedValue(entry.diff) {
							id = entry.decreasedId
						}
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, entry.field)
						result = append(result, NewApiChange(
							id,
							config,
							[]any{paramLocation, paramName, entry.diff.From, entry.diff.To},
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
	}
	return result
}
