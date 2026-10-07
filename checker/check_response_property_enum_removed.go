package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyEnumRemovedId = "response-property-enum-removed"
)

func ResponsePropertyEnumRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if !enumRemoved(p.propertyDiff) {
				return
			}

			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "enum")
			result = append(result, p.newChange(
				ResponsePropertyEnumRemovedId,
				[]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName), info.responseStatus},
				commentId(ResponsePropertyEnumRemovedId),
			).WithSources(baseSource, revisionSource))
		})
	})

	return result
}
