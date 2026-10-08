package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyBecameNotEnumId = "request-property-became-not-enum"
)

func RequestPropertyBecameNotEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if !enumRemoved(p.propertyDiff) {
				return
			}

			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "enum")
			result = append(result, p.newChange(
				RequestPropertyBecameNotEnumId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)},
				"",
			).WithSources(baseSource, revisionSource))
		})
	})

	return result
}
