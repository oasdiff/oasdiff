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
			id    string
			field string
			bound diff.SchemaBound
		}{
			{RequestParameterMaxSetId, "maximum", maximumBound},
			{RequestParameterExclusiveMaxSetId, "exclusiveMaximum", exclusiveMaximumBound},
		} {
			value, ok := entry.bound.WasSet(p.paramDiff.SchemaDiff)
			if !ok {
				continue
			}
			_, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, entry.field)
			result = append(result, p.opInfo.NewApiChange(
				entry.id,
				[]any{p.location, p.name, value},
				boundSetComment,
			).WithSources(nil, revisionSource))
		}
	})
	return result
}
