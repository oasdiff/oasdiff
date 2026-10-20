package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyBecameEnumId = "request-property-became-enum"
)

func RequestPropertyBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if enumDiff := p.propertyDiff.EnumDiff; enumDiff == nil || !enumDiff.EnumAdded {
				return
			}

			propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "enum")
			result = append(result, p.newChange(
				RequestPropertyBecameEnumId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)},
				"",
			).WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
