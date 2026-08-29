package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyUniqueItemsSetId       = "request-body-unique-items-set"
	RequestPropertyUniqueItemsSetId   = "request-property-unique-items-set"
	RequestBodyUniqueItemsUnsetId     = "request-body-unique-items-unset"
	RequestPropertyUniqueItemsUnsetId = "request-property-unique-items-unset"
)

func RequestPropertyUniqueItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if uniqueItemsDiff := info.schemaDiff.UniqueItemsDiff; uniqueItemsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "uniqueItems")
			if uniqueItemsDiff.To == true {
				result = append(result, info.newChange(
					RequestBodyUniqueItemsSetId,
					[]any{},
					"",
				).WithSources(baseSource, revisionSource))
			} else {
				result = append(result, info.newChange(
					RequestBodyUniqueItemsUnsetId,
					[]any{},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			uniqueItemsDiff := p.propertyDiff.UniqueItemsDiff
			if uniqueItemsDiff == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "uniqueItems")

			if uniqueItemsDiff.To == true {
				if p.propertyDiff.Revision.ReadOnly {
					return
				}
				result = append(result, p.newChange(
					RequestPropertyUniqueItemsSetId,
					[]any{propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			} else {
				result = append(result, p.newChange(
					RequestPropertyUniqueItemsUnsetId,
					[]any{propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
