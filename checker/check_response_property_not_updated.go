package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyNotAddedId       = "response-body-not-added"
	ResponseBodyNotRemovedId     = "response-body-not-removed"
	ResponsePropertyNotAddedId   = "response-property-not-added"
	ResponsePropertyNotRemovedId = "response-property-not-removed"
)

func ResponsePropertyNotUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.NotDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "not")
			if info.schemaDiff.NotDiff.SchemaAdded {
				result = append(result, info.newChange(ResponseBodyNotAddedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.NotDiff.SchemaDeleted {
				result = append(result, info.newChange(ResponseBodyNotRemovedId, []any{info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.NotDiff == nil {
				return
			}
			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "not")
			if p.propertyDiff.NotDiff.SchemaAdded {
				result = append(result, p.newChange(ResponsePropertyNotAddedId, []any{propName, info.responseStatus}, "").
					WithSources(nil, propRevisionSource))
			}
			if p.propertyDiff.NotDiff.SchemaDeleted {
				result = append(result, p.newChange(ResponsePropertyNotRemovedId, []any{propName, info.responseStatus}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
