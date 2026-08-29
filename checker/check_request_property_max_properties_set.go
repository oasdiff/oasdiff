package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxPropertiesSetId     = "request-body-max-properties-set"
	RequestPropertyMaxPropertiesSetId = "request-property-max-properties-set"
)

func RequestPropertyMaxPropertiesSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxPropertiesDiff := info.schemaDiff.MaxPropsDiff; maxPropertiesDiff != nil &&
			maxPropertiesDiff.From == nil && maxPropertiesDiff.To != nil {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxProperties")
			result = append(result, info.newChange(
				RequestBodyMaxPropertiesSetId,
				[]any{maxPropertiesDiff.To},
				commentId(RequestBodyMaxPropertiesSetId),
			).WithSources(nil, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxPropertiesDiff := p.propertyDiff.MaxPropsDiff
			if maxPropertiesDiff == nil || maxPropertiesDiff.From != nil || maxPropertiesDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}

			_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxProperties")
			result = append(result, p.newChange(
				RequestPropertyMaxPropertiesSetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxPropertiesDiff.To},
				commentId(RequestPropertyMaxPropertiesSetId),
			).WithSources(nil, propRevisionSource))
		})
	})

	return result
}
