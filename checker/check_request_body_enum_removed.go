package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyEnumRemovedId = "request-body-enum-removed"
)

func RequestBodyEnumRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if !enumRemoved(info.schemaDiff) {
			return
		}
		baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "enum")
		result = append(result, info.newChange(RequestBodyEnumRemovedId, nil, "").
			WithSources(baseSource, revisionSource))
	})

	return result
}
