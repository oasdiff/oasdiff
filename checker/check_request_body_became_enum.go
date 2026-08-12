package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecameEnumId = "request-body-became-enum"
)

func RequestBodyBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.EnumDiff == nil || !info.schemaDiff.EnumDiff.EnumAdded {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "enum")
		result = append(result, info.newChange(RequestBodyBecameEnumId, nil, "").
			WithSources(baseSource, revisionSource))
	})

	return result
}
