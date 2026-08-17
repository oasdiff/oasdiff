package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxItemsIncreasedId     = "response-body-max-items-increased"
	ResponsePropertyMaxItemsIncreasedId = "response-property-max-items-increased"
)

func ResponsePropertyMaxItemsIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxItemsDiff := info.schemaDiff.MaxItemsDiff; maxItemsDiff != nil &&
			maxItemsDiff.From != nil &&
			maxItemsDiff.To != nil &&
			isIncreasedValue(maxItemsDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxItems")
			result = append(result, info.newChange(
				ResponseBodyMaxItemsIncreasedId,
				[]any{maxItemsDiff.From, maxItemsDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxItemsDiff := p.propertyDiff.MaxItemsDiff
			if maxItemsDiff == nil {
				return
			}
			if maxItemsDiff.To == nil ||
				maxItemsDiff.From == nil {
				return
			}
			if !isIncreasedValue(maxItemsDiff) {
				return
			}

			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxItems")
			result = append(result, p.newChange(
				ResponsePropertyMaxItemsIncreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxItemsDiff.From, maxItemsDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
