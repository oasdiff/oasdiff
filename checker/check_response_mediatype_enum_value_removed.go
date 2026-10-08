package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeEnumValueRemovedId = "response-mediatype-enum-value-removed"
)

func ResponseMediaTypeEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		for _, enumVal := range deletedEnumValues(info.schemaDiff) {
			baseSource, revisionSource := location.SchemaDeletedItemSources(operationsSources, info.operationItem, info.schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
			result = append(result, info.newChange(ResponseMediaTypeEnumValueRemovedId, []any{info.mediaType, enumVal}, "").
				WithSources(baseSource, revisionSource))
		}
	})

	return result
}
