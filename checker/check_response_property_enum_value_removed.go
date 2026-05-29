package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyEnumValueRemovedId = "response-property-enum-value-removed"
)

func ResponseParameterEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			enumDiff := p.propertyDiff.EnumDiff
			if enumDiff == nil || enumDiff.Deleted == nil {
				return
			}

			for _, enumVal := range enumDiff.Deleted {
				baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, info.operationItem, p.propertyDiff, "enum", fmt.Sprintf("%v", enumVal))
				result = append(result, p.newChange(
					ResponsePropertyEnumValueRemovedId,
					[]any{enumVal, propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					"",
				).WithSources(baseSource, revisionSource))
			}
		})
	})

	return result
}
