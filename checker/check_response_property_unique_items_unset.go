package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyUniqueItemsUnsetId     = "response-body-unique-items-unset"
	ResponsePropertyUniqueItemsUnsetId = "response-property-unique-items-unset"
)

func ResponsePropertyUniqueItemsUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if uniqueItemsDiff := info.schemaDiff.UniqueItemsDiff; uniqueItemsDiff != nil &&
			uniqueItemsDiff.To == false {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "uniqueItems")
			result = append(result, info.newChange(
				ResponseBodyUniqueItemsUnsetId,
				[]any{},
				"",
			).WithSources(baseSource, revisionSource))
		}

		info.walkProperties(func(p propertyInfo) {
			uniqueItemsDiff := p.propertyDiff.UniqueItemsDiff
			if uniqueItemsDiff == nil || uniqueItemsDiff.To != false {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "uniqueItems")
			result = append(result, p.newChange(
				ResponsePropertyUniqueItemsUnsetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
