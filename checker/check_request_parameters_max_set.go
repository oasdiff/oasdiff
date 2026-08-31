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
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		for _, entry := range []struct {
			diff  *diff.ValueDiff
			id    string
			field string
		}{
			{p.paramDiff.SchemaDiff.MaxDiff, RequestParameterMaxSetId, "maximum"},
			{p.paramDiff.SchemaDiff.ExclusiveMaxDiff, RequestParameterExclusiveMaxSetId, "exclusiveMaximum"},
		} {
			if entry.diff == nil || entry.diff.From != nil || entry.diff.To == nil {
				continue
			}
			_, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, entry.field)
			result = append(result, p.opInfo.NewApiChange(
				entry.id,
				[]any{p.location, p.name, entry.diff.To},
				commentId(entry.id),
			).WithSources(nil, revisionSource))
		}
	})
	return result
}
