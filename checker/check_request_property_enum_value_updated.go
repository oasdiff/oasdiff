package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyEnumValueRemovedId         = "request-property-enum-value-removed"
	RequestReadOnlyPropertyEnumValueRemovedId = "request-read-only-property-enum-value-removed"
	RequestPropertyEnumValueAddedId           = "request-property-enum-value-added"
)

func RequestPropertyEnumValueUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			enumDiff := p.propertyDiff.EnumDiff
			if enumDiff == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)

			for _, enumVal := range enumDiff.Deleted {
				baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, info.operationItem, p.propertyDiff, "enum", fmt.Sprintf("%v", enumVal))

				id := RequestPropertyEnumValueRemovedId
				if p.propertyDiff.Revision.ReadOnly {
					id = RequestReadOnlyPropertyEnumValueRemovedId
				}

				result = append(result, p.newChange(
					id,
					[]any{enumVal, propName},
					"",
				).WithSources(baseSource, revisionSource))
			}

			for _, enumVal := range enumDiff.Added {
				baseSource, revisionSource := SchemaAddedItemSources(operationsSources, info.operationItem, p.propertyDiff, "enum", fmt.Sprintf("%v", enumVal))
				result = append(result, p.newChange(
					RequestPropertyEnumValueAddedId,
					[]any{enumVal, propName},
					"",
				).WithSources(baseSource, revisionSource))
			}
		})
	})

	return result
}
