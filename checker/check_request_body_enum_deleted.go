package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyEnumValueRemovedId = "request-body-enum-value-removed"
)

func RequestBodyEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		for _, enumVal := range deletedEnumValues(info.schemaDiff) {
			baseSource, revisionSource := location.SchemaDeletedItemSources(operationsSources, info.operationItem, info.schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
			result = append(result, info.newChange(RequestBodyEnumValueRemovedId, []any{enumVal}, "").
				WithSources(baseSource, revisionSource))
		}
	})

	return result
}
