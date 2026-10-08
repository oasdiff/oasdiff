package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeBecameEnumId = "response-mediatype-became-enum"
)

func ResponseMediaTypeBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if !enumAdded(info.schemaDiff) {
			return
		}
		baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "enum")
		result = append(result, info.newChange(ResponseMediaTypeBecameEnumId, []any{info.mediaType, info.responseStatus}, "").
			WithSources(baseSource, revisionSource))
	})

	return result
}
