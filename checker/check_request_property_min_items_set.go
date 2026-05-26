package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMinItemsSetId     = "request-body-min-items-set"
	RequestPropertyMinItemsSetId = "request-property-min-items-set"
)

func RequestPropertyMinItemsSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minItemsDiff := info.schemaDiff.MinItemsDiff; minItemsDiff != nil &&
			minItemsDiff.From == nil && minItemsDiff.To != nil {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minItems")
			result = append(result, info.newChange(
				RequestBodyMinItemsSetId,
				[]any{minItemsDiff.To},
				commentId(RequestBodyMinItemsSetId),
			).WithSources(nil, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			minItemsDiff := p.propertyDiff.MinItemsDiff
			if minItemsDiff == nil || minItemsDiff.From != nil || minItemsDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}

			_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minItems")
			result = append(result, p.newChange(
				RequestPropertyMinItemsSetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minItemsDiff.To},
				commentId(RequestPropertyMinItemsSetId),
			).WithSources(nil, propRevisionSource))
		})
	})

	return result
}
