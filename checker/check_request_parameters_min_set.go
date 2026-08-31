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
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		for _, entry := range []struct {
			diff  *diff.ValueDiff
			id    string
			field string
		}{
			{p.paramDiff.SchemaDiff.MinDiff, RequestParameterMinSetId, "minimum"},
			{p.paramDiff.SchemaDiff.ExclusiveMinDiff, RequestParameterExclusiveMinSetId, "exclusiveMinimum"},
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
