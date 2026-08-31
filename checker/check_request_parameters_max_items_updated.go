package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterMaxItemsIncreasedId = "request-parameter-max-items-increased"
	RequestParameterMaxItemsDecreasedId = "request-parameter-max-items-decreased"
)

func RequestParameterMaxItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "maxItems")

		// Check for maxItems on the parameter schema itself (for array parameters)
		maxItemsDiff := p.paramDiff.SchemaDiff.MaxItemsDiff
		if maxItemsDiff == nil && p.paramDiff.SchemaDiff.ItemsDiff != nil {
			// Fallback: check for maxItems on the items schema (legacy behavior)
			maxItemsDiff = p.paramDiff.SchemaDiff.ItemsDiff.MaxItemsDiff
		}

		if maxItemsDiff == nil {
			return
		}
		if maxItemsDiff.From == nil ||
			maxItemsDiff.To == nil {
			return
		}

		id := RequestParameterMaxItemsDecreasedId
		if isIncreasedValue(maxItemsDiff) {
			id = RequestParameterMaxItemsIncreasedId
		}

		result = append(result, p.opInfo.NewApiChange(
			id,
			[]any{p.location, p.name, maxItemsDiff.From, maxItemsDiff.To},
			"",
		).WithSources(baseSource, revisionSource))
	})
	return result
}
