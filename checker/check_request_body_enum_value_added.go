package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyEnumValueAddedId = "request-body-enum-value-added"
)

func RequestBodyEnumValueAddedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.EnumDiff == nil {
			return
		}
		for _, enumVal := range addedEnumValues(info.schemaDiff) {
			baseSource, revisionSource := location.SchemaAddedItemSources(operationsSources, info.operationItem, info.schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
			result = append(result, info.newChange(RequestBodyEnumValueAddedId, []any{enumVal}, "").
				WithSources(baseSource, revisionSource))
		}
	})

	return result
}
