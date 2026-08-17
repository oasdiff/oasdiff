package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinPropertiesDecreasedId     = "response-body-min-properties-decreased"
	ResponsePropertyMinPropertiesDecreasedId = "response-property-min-properties-decreased"
)

func ResponsePropertyMinPropertiesDecreasedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minPropertiesDiff := info.schemaDiff.MinPropsDiff; minPropertiesDiff != nil &&
			minPropertiesDiff.From != nil && minPropertiesDiff.To != nil && isDecreasedValue(minPropertiesDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minProperties")
			result = append(result, info.newChange(
				ResponseBodyMinPropertiesDecreasedId,
				[]any{minPropertiesDiff.From, minPropertiesDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minPropertiesDiff := p.propertyDiff.MinPropsDiff
			if minPropertiesDiff == nil || minPropertiesDiff.To == nil || minPropertiesDiff.From == nil {
				return
			}
			if !isDecreasedValue(minPropertiesDiff) {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minProperties")
			result = append(result, p.newChange(
				ResponsePropertyMinPropertiesDecreasedId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minPropertiesDiff.From, minPropertiesDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
