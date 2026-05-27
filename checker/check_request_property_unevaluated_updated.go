package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyUnevaluatedItemsAddedId        = "request-body-unevaluated-items-added"
	RequestBodyUnevaluatedItemsRemovedId      = "request-body-unevaluated-items-removed"
	RequestBodyUnevaluatedPropertiesAddedId   = "request-body-unevaluated-properties-added"
	RequestBodyUnevaluatedPropertiesRemovedId = "request-body-unevaluated-properties-removed"

	RequestPropertyUnevaluatedItemsAddedId        = "request-property-unevaluated-items-added"
	RequestPropertyUnevaluatedItemsRemovedId      = "request-property-unevaluated-items-removed"
	RequestPropertyUnevaluatedPropertiesAddedId   = "request-property-unevaluated-properties-added"
	RequestPropertyUnevaluatedPropertiesRemovedId = "request-property-unevaluated-properties-removed"
)

func RequestPropertyUnevaluatedUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.UnevaluatedItemsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedItems")
			if info.schemaDiff.UnevaluatedItemsDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyUnevaluatedItemsAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.UnevaluatedItemsDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyUnevaluatedItemsRemovedId, nil, "").
					WithSources(baseSource, nil))
			}
		}

		if info.schemaDiff.UnevaluatedPropertiesDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "unevaluatedProperties")
			if info.schemaDiff.UnevaluatedPropertiesDiff.SchemaAdded {
				result = append(result, info.newChange(RequestBodyUnevaluatedPropertiesAddedId, nil, "").
					WithSources(nil, revisionSource))
			}
			if info.schemaDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
				result = append(result, info.newChange(RequestBodyUnevaluatedPropertiesRemovedId, nil, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			propName := propertyFullName(p.propertyPath, p.propertyName)

			if p.propertyDiff.UnevaluatedItemsDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedItems")
				if p.propertyDiff.UnevaluatedItemsDiff.SchemaAdded {
					result = append(result, p.newChange(RequestPropertyUnevaluatedItemsAddedId, []any{propName}, "").
						WithSources(nil, propRevisionSource))
				}
				if p.propertyDiff.UnevaluatedItemsDiff.SchemaDeleted {
					result = append(result, p.newChange(RequestPropertyUnevaluatedItemsRemovedId, []any{propName}, "").
						WithSources(propBaseSource, nil))
				}
			}

			if p.propertyDiff.UnevaluatedPropertiesDiff != nil {
				propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "unevaluatedProperties")
				if p.propertyDiff.UnevaluatedPropertiesDiff.SchemaAdded {
					result = append(result, p.newChange(RequestPropertyUnevaluatedPropertiesAddedId, []any{propName}, "").
						WithSources(nil, propRevisionSource))
				}
				if p.propertyDiff.UnevaluatedPropertiesDiff.SchemaDeleted {
					result = append(result, p.newChange(RequestPropertyUnevaluatedPropertiesRemovedId, []any{propName}, "").
						WithSources(propBaseSource, nil))
				}
			}
		})
	})

	return result
}
