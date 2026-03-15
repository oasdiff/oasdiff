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
					for _, entry := range []struct {
						diff        *diff.ValueDiff
						decreasedId string
						increasedId string
						field       string
					}{
						{paramDiff.SchemaDiff.MaxDiff, RequestParameterMaxDecreasedId, RequestParameterMaxIncreasedId, "maximum"},
						{paramDiff.SchemaDiff.ExclusiveMaxDiff, RequestParameterExclusiveMaxDecreasedId, RequestParameterExclusiveMaxIncreasedId, "exclusiveMaximum"},
					} {
						if entry.diff == nil || entry.diff.From == nil || entry.diff.To == nil {
							continue
						}
						id := entry.decreasedId
						if !IsDecreasedValue(entry.diff) {
							id = entry.increasedId
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
