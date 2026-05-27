package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinItemsDecreasedId     = "response-body-min-items-decreased"
	ResponsePropertyMinItemsDecreasedId = "response-property-min-items-decreased"
)

func ResponsePropertyMinItemsDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minItemsDiff := info.schemaDiff.MinItemsDiff; minItemsDiff != nil &&
			minItemsDiff.From != nil && minItemsDiff.To != nil && IsDecreasedValue(minItemsDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minItems")
			result = append(result, info.newChange(
				ResponseBodyMinItemsDecreasedId,
				[]any{minItemsDiff.From, minItemsDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minItemsDiff := p.propertyDiff.MinItemsDiff
			if minItemsDiff == nil || minItemsDiff.To == nil || minItemsDiff.From == nil {
				return
			}
			if !IsDecreasedValue(minItemsDiff) {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minItems")
			result = append(result, p.newChange(
				ResponsePropertyMinItemsDecreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minItemsDiff.From, minItemsDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
