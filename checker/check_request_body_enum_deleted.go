package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyEnumValueRemovedId = "request-body-enum-value-removed"
)

func RequestBodyEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		enumDiff := info.schemaDiff.EnumDiff
		if enumDiff == nil || enumDiff.Deleted == nil {
			return
		}
		for _, enumVal := range enumDiff.Deleted {
			baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, info.operationItem, info.schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
			result = append(result, info.newChange(RequestBodyEnumValueRemovedId, []any{enumVal}, "").
				WithSources(baseSource, revisionSource))
		}
	})

	return result
}
