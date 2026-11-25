package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyNotAddedId       = "request-body-not-added"
	RequestBodyNotRemovedId     = "request-body-not-removed"
	RequestPropertyNotAddedId   = "request-property-not-added"
	RequestPropertyNotRemovedId = "request-property-not-removed"
)

func RequestPropertyNotUpdatedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.NotDiff != nil {
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "not")
			if info.schemaDiff.NotDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyNotAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.NotDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyNotRemovedId, nil, "").
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
				result = append(result, p.newChange(RequestPropertyNotAddedId, []any{propName}, "").
					WithSources(nil, propRevisionSource))
			}
			if p.propertyDiff.NotDiff.SchemaDeleted {
				result = append(result, p.newChange(RequestPropertyNotRemovedId, []any{propName}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
