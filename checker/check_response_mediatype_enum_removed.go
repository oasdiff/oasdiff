package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeEnumRemovedId = "response-mediatype-enum-removed"
)

func ResponseMediaTypeEnumRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if !enumRemoved(info.schemaDiff) {
			return
		}
		baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "enum")
		result = append(result, info.newChange(ResponseMediaTypeEnumRemovedId, []any{info.mediaType, info.responseStatus}, commentId(ResponseMediaTypeEnumRemovedId)).
			WithSources(baseSource, revisionSource))
	})

	return result
}
