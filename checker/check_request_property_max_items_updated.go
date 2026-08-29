package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxItemsDecreasedId             = "request-body-max-items-decreased"
	RequestBodyMaxItemsIncreasedId             = "request-body-max-items-increased"
	RequestPropertyMaxItemsDecreasedId         = "request-property-max-items-decreased"
	RequestReadOnlyPropertyMaxItemsDecreasedId = "request-read-only-property-max-items-decreased"
	RequestPropertyMaxItemsIncreasedId         = "request-property-max-items-increased"
)

func RequestPropertyMaxItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxItems")
		if maxItemsDiff := info.schemaDiff.MaxItemsDiff; maxItemsDiff != nil &&
			maxItemsDiff.From != nil &&
			maxItemsDiff.To != nil {
			if isDecreasedValue(maxItemsDiff) {
				result = append(result, info.newChange(
					RequestBodyMaxItemsDecreasedId,
					[]any{maxItemsDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyMaxItemsIncreasedId,
					[]any{maxItemsDiff.From, maxItemsDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			maxItemsDiff := p.propertyDiff.MaxItemsDiff
			if maxItemsDiff == nil {
				return
			}
			if maxItemsDiff.From == nil ||
				maxItemsDiff.To == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxItems")

			if isDecreasedValue(maxItemsDiff) {

				id := RequestPropertyMaxItemsDecreasedId

				if p.propertyDiff.Revision.ReadOnly {
					id = RequestReadOnlyPropertyMaxItemsDecreasedId
				}

				result = append(result, p.newChange(
					id,
					[]any{propName, maxItemsDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			} else {
				result = append(result, p.newChange(
					RequestPropertyMaxItemsIncreasedId,
					[]any{propName, maxItemsDiff.From, maxItemsDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
