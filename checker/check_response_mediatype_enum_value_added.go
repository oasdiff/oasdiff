package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeEnumValueAddedId = "response-mediatype-enum-value-added"
)

func ResponseMediaTypeEnumValueAddedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		for _, enumVal := range addedEnumValues(info.schemaDiff) {
			baseSource, revisionSource := location.SchemaAddedItemSources(operationsSources, info.operationItem, info.schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
			result = append(result, info.newChange(ResponseMediaTypeEnumValueAddedId, []any{enumVal, info.mediaType, info.responseStatus}, commentId(ResponseMediaTypeEnumValueAddedId)).
				WithSources(baseSource, revisionSource))
		}
	})

	return result
}
