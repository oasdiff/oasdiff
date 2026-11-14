package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameNotEnumId = "response-property-became-not-enum"
)

func ResponsePropertyBecameNotEnumCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if !becameNotEnum(p.propertyDiff) {
				return
			}

			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "enum")
			result = append(result, p.newChange(
				ResponsePropertyBecameNotEnumId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
				commentId(ResponsePropertyBecameNotEnumId),
			).WithSources(baseSource, revisionSource))
		})
	})

	return result
}
