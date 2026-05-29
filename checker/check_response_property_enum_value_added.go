package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyEnumValueAddedId          = "response-property-enum-value-added"
	ResponseWriteOnlyPropertyEnumValueAddedId = "response-write-only-property-enum-value-added"
)

func ResponsePropertyEnumValueAddedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			enumDiff := p.propertyDiff.EnumDiff
			if enumDiff == nil || enumDiff.Added == nil {
				return
			}

			id := ResponsePropertyEnumValueAddedId
			comment := commentId(ResponsePropertyEnumValueAddedId)

			if p.propertyDiff.Revision.WriteOnly {
				// Document write-only enum update
				id = ResponseWriteOnlyPropertyEnumValueAddedId
				comment = ""
			}

			for _, enumVal := range enumDiff.Added {
				baseSource, revisionSource := SchemaAddedItemSources(operationsSources, info.operationItem, p.propertyDiff, "enum", fmt.Sprintf("%v", enumVal))
				result = append(result, p.newChange(
					id,
					[]any{enumVal, propertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
					comment,
				).WithSources(baseSource, revisionSource))
			}
		})
	})

	return result
}
