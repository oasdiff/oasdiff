package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
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
			!uintBoundUnset(minPropertiesDiff) && isDecreasedValue(minPropertiesDiff) {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minProperties")
			result = append(result, info.newChange(
				ResponseBodyMinPropertiesDecreasedId,
				[]any{minPropertiesDiff.From, minPropertiesDiff.To},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minPropertiesDiff := p.propertyDiff.MinPropsDiff
			if minPropertiesDiff == nil || uintBoundUnset(minPropertiesDiff) {
				return
			}
			if !isDecreasedValue(minPropertiesDiff) {
				return
			}

			propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minProperties")
			result = append(result, p.newChange(
				ResponsePropertyMinPropertiesDecreasedId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName), minPropertiesDiff.From, minPropertiesDiff.To, info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
