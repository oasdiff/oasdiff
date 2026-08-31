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
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		for _, entry := range []struct {
			diff        *diff.ValueDiff
			decreasedId string
			increasedId string
			field       string
		}{
			{p.paramDiff.SchemaDiff.MaxDiff, RequestParameterMaxDecreasedId, RequestParameterMaxIncreasedId, "maximum"},
			{p.paramDiff.SchemaDiff.ExclusiveMaxDiff, RequestParameterExclusiveMaxDecreasedId, RequestParameterExclusiveMaxIncreasedId, "exclusiveMaximum"},
		} {
			if entry.diff == nil || entry.diff.From == nil || entry.diff.To == nil {
				continue
			}
			id := entry.decreasedId
			if !isDecreasedValue(entry.diff) {
				id = entry.increasedId
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
