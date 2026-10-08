package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecameNotEnumId = "request-body-became-not-enum"
)

func RequestBodyBecameNotEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if !enumRemoved(info.schemaDiff) {
			return
		}
		baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "enum")
		result = append(result, info.newChange(RequestBodyBecameNotEnumId, nil, "").
			WithSources(baseSource, revisionSource))
	})

	return result
}
