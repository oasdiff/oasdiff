package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMaxItemsSetId     = "request-body-max-items-set"
	RequestPropertyMaxItemsSetId = "request-property-max-items-set"
)

func RequestPropertyMaxItemsSetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if maxItemsDiff := info.schemaDiff.MaxItemsDiff; maxItemsDiff != nil &&
			maxItemsDiff.From == nil && maxItemsDiff.To != nil {
			_, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "maxItems")
			result = append(result, info.newChange(
				RequestBodyMaxItemsSetId,
				[]any{maxItemsDiff.To},
				commentId(RequestBodyMaxItemsSetId),
			).WithSources(nil, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			maxItemsDiff := p.propertyDiff.MaxItemsDiff
			if maxItemsDiff == nil || maxItemsDiff.From != nil || maxItemsDiff.To == nil {
				return
			}
			if p.propertyDiff.Revision.ReadOnly {
				return
			}

			_, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "maxItems")
			result = append(result, p.newChange(
				RequestPropertyMaxItemsSetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), maxItemsDiff.To},
				commentId(RequestPropertyMaxItemsSetId),
			).WithSources(nil, propRevisionSource))
		})
	})

	return result
}
