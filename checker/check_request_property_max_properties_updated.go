package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxPropertiesDecreasedId             = "request-body-max-properties-decreased"
	RequestBodyMaxPropertiesIncreasedId             = "request-body-max-properties-increased"
	RequestPropertyMaxPropertiesDecreasedId         = "request-property-max-properties-decreased"
	RequestReadOnlyPropertyMaxPropertiesDecreasedId = "request-read-only-property-max-properties-decreased"
	RequestPropertyMaxPropertiesIncreasedId         = "request-property-max-properties-increased"
)

func RequestPropertyMaxPropertiesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxProperties")
		if maxPropertiesDiff := info.schemaDiff.MaxPropsDiff; maxPropertiesDiff != nil &&
			maxPropertiesDiff.From != nil &&
			maxPropertiesDiff.To != nil {
			if isDecreasedValue(maxPropertiesDiff) {
				result = append(result, info.newChange(
					RequestBodyMaxPropertiesDecreasedId,
					[]any{maxPropertiesDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyMaxPropertiesIncreasedId,
					[]any{maxPropertiesDiff.From, maxPropertiesDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			maxPropertiesDiff := p.propertyDiff.MaxPropsDiff
			if maxPropertiesDiff == nil {
				return
			}
			if maxPropertiesDiff.From == nil ||
				maxPropertiesDiff.To == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxProperties")

			if isDecreasedValue(maxPropertiesDiff) {

				id := RequestPropertyMaxPropertiesDecreasedId

				if p.propertyDiff.Revision.ReadOnly {
					id = RequestReadOnlyPropertyMaxPropertiesDecreasedId
				}

				result = append(result, p.newChange(
					id,
					[]any{propName, maxPropertiesDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			} else {
				result = append(result, p.newChange(
					RequestPropertyMaxPropertiesIncreasedId,
					[]any{propName, maxPropertiesDiff.From, maxPropertiesDiff.To},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
