package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyUnevaluatedItemsAddedId        = "response-body-unevaluated-items-added"
	ResponseBodyUnevaluatedItemsRemovedId      = "response-body-unevaluated-items-removed"
	ResponseBodyUnevaluatedPropertiesAddedId   = "response-body-unevaluated-properties-added"
	ResponseBodyUnevaluatedPropertiesRemovedId = "response-body-unevaluated-properties-removed"

	ResponsePropertyUnevaluatedItemsAddedId        = "response-property-unevaluated-items-added"
	ResponsePropertyUnevaluatedItemsRemovedId      = "response-property-unevaluated-items-removed"
	ResponsePropertyUnevaluatedPropertiesAddedId   = "response-property-unevaluated-properties-added"
	ResponsePropertyUnevaluatedPropertiesRemovedId = "response-property-unevaluated-properties-removed"
)

func ResponsePropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.UnevaluatedItemsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedItems")
			if info.schemaDiff.UnevaluatedItemsDiff.SchemaAdded {
				result = append(result, info.newChange(ResponseBodyUnevaluatedItemsAddedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
				result = append(result, info.newChange(ResponseBodyUnevaluatedItemsRemovedId, []any{info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		}

		if info.schemaDiff.UnevaluatedPropertiesDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedProperties")
			if info.schemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
				result = append(result, info.newChange(ResponseBodyUnevaluatedPropertiesAddedId, []any{info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
				result = append(result, info.newChange(ResponseBodyUnevaluatedPropertiesRemovedId, []any{info.responseStatus}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if p.propertyDiff.UnevaluatedItemsDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedItems")
				if p.propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
					result = append(result, p.newChange(ResponsePropertyUnevaluatedItemsAddedId, []any{propName, info.responseStatus}, "").
						WithSources(nil, propRevisionSource))
				}
				if p.propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
					result = append(result, p.newChange(ResponsePropertyUnevaluatedItemsRemovedId, []any{propName, info.responseStatus}, "").
						WithSources(propBaseSource, nil))
				}
			}

			if p.propertyDiff.UnevaluatedPropertiesDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedProperties")
				if p.propertyDiff.UnevaluatedPropertiesDiff.SchemaAdded {
					result = append(result, p.newChange(ResponsePropertyUnevaluatedPropertiesAddedId, []any{propName, info.responseStatus}, "").
						WithSources(nil, propRevisionSource))
				}
				if p.propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
					result = append(result, p.newChange(ResponsePropertyUnevaluatedPropertiesRemovedId, []any{propName, info.responseStatus}, "").
						WithSources(propBaseSource, nil))
				}
			}
		})
	})

	return result
}
