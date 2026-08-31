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
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		for _, entry := range []struct {
			diff        *diff.ValueDiff
			increasedId string
			decreasedId string
			field       string
		}{
			{p.paramDiff.SchemaDiff.MinDiff, RequestParameterMinIncreasedId, RequestParameterMinDecreasedId, "minimum"},
			{p.paramDiff.SchemaDiff.ExclusiveMinDiff, RequestParameterExclusiveMinIncreasedId, RequestParameterExclusiveMinDecreasedId, "exclusiveMinimum"},
		} {
			if entry.diff == nil || entry.diff.From == nil || entry.diff.To == nil {
				continue
			}
			id := entry.increasedId
			if !isIncreasedValue(entry.diff) {
				id = entry.decreasedId
			}
			baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, entry.field)
			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name, entry.diff.From, entry.diff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}
	})
	return result
}
