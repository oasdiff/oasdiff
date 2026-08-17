package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMaxPropertiesIncreasedId     = "response-body-max-properties-increased"
	ResponsePropertyMaxPropertiesIncreasedId = "response-property-max-properties-increased"
)

func ResponsePropertyMaxPropertiesIncreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxPropertiesDiff := info.schemaDiff.MaxPropsDiff; maxPropertiesDiff != nil &&
			maxPropertiesDiff.From != nil &&
			maxPropertiesDiff.To != nil &&
			isIncreasedValue(maxPropertiesDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxProperties")
			result = append(result, info.newChange(
				ResponseBodyMaxPropertiesIncreasedId,
				[]any{maxPropertiesDiff.From, maxPropertiesDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxPropertiesDiff := p.propertyDiff.MaxPropsDiff
			if maxPropertiesDiff == nil {
				return
			}
			if maxPropertiesDiff.To == nil ||
				maxPropertiesDiff.From == nil {
				return
			}
			if !isIncreasedValue(maxPropertiesDiff) {
				return
			}

			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxProperties")
			result = append(result, p.newChange(
				ResponsePropertyMaxPropertiesIncreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxPropertiesDiff.From, maxPropertiesDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
