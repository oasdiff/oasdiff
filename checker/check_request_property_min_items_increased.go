package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinItemsIncreasedId     = "request-body-min-items-increased"
	RequestPropertyMinItemsIncreasedId = "request-property-min-items-increased"
)

func RequestPropertyMinItemsIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minItemsDiff := info.schemaDiff.MinItemsDiff; minItemsDiff != nil &&
			minItemsDiff.From != nil && minItemsDiff.To != nil && IsIncreasedValue(minItemsDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minItems")
			result = append(result, info.newChange(
				RequestBodyMinItemsIncreasedId,
				[]any{minItemsDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minItemsDiff := p.propertyDiff.MinItemsDiff
			if minItemsDiff == nil || minItemsDiff.From == nil || minItemsDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}
			if !IsIncreasedValue(minItemsDiff) {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minItems")
			result = append(result, p.newChange(
				RequestPropertyMinItemsIncreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minItemsDiff.To},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
