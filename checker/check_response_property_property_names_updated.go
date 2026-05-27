package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPropertyNamesAddedId       = "response-body-property-names-added"
	ResponseBodyPropertyNamesRemovedId     = "response-body-property-names-removed"
	ResponsePropertyPropertyNamesAddedId   = "response-property-property-names-added"
	ResponsePropertyPropertyNamesRemovedId = "response-property-property-names-removed"
)

func ResponsePropertyPropertyNamesUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PropertyNamesDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "propertyNames")
			if info.schemaDiff.PropertyNamesDiff.SchemaAdded {
				result = append(result, info.newChange(ResponseBodyPropertyNamesAddedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.PropertyNamesDiff.SchemaDeleted {
				result = append(result, info.newChange(ResponseBodyPropertyNamesRemovedId, []any{info.responseStatus}, "").
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
				result = append(result, p.newChange(ResponsePropertyPropertyNamesAddedId, []any{propName, info.responseStatus}, "").
					WithSources(nil, propRevisionSource))
			}
			if p.propertyDiff.PropertyNamesDiff.SchemaDeleted {
				result = append(result, p.newChange(ResponsePropertyPropertyNamesRemovedId, []any{propName, info.responseStatus}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
