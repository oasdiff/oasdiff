package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinPropertiesIncreasedId     = "request-body-min-properties-increased"
	RequestPropertyMinPropertiesIncreasedId = "request-property-min-properties-increased"
)

func RequestPropertyMinPropertiesIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minPropertiesDiff := info.schemaDiff.MinPropsDiff; minPropertiesDiff != nil &&
			minPropertiesDiff.From != nil && minPropertiesDiff.To != nil && isIncreasedValue(minPropertiesDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minProperties")
			result = append(result, info.newChange(
				RequestBodyMinPropertiesIncreasedId,
				[]any{minPropertiesDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minPropertiesDiff := p.propertyDiff.MinPropsDiff
			if minPropertiesDiff == nil || minPropertiesDiff.From == nil || minPropertiesDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}
			if !isIncreasedValue(minPropertiesDiff) {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minProperties")
			result = append(result, p.newChange(
				RequestPropertyMinPropertiesIncreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minPropertiesDiff.To},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
