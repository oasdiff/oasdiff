package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPropertyNamesAddedId       = "request-body-property-names-added"
	RequestBodyPropertyNamesRemovedId     = "request-body-property-names-removed"
	RequestPropertyPropertyNamesAddedId   = "request-property-property-names-added"
	RequestPropertyPropertyNamesRemovedId = "request-property-property-names-removed"
)

func RequestPropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PropertyNamesDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "propertyNames")
			if info.schemaDiff.PropertyNamesDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyPropertyNamesAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.PropertyNamesDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyPropertyNamesRemovedId, nil, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.PropertyNamesDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "propertyNames")
			if p.propertyDiff.PropertyNamesDiff.SchemaAdded {
				result = append(result, p.newChange(RequestPropertyPropertyNamesAddedId, []any{propName}, "").
					WithSources(nil, propRevisionSource))
			}
			if p.propertyDiff.PropertyNamesDiff.SchemaDeleted {
				result = append(result, p.newChange(RequestPropertyPropertyNamesRemovedId, []any{propName}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
