package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPrefixItemsAddedId       = "request-body-prefix-items-added"
	RequestBodyPrefixItemsRemovedId     = "request-body-prefix-items-removed"
	RequestPropertyPrefixItemsAddedId   = "request-property-prefix-items-added"
	RequestPropertyPrefixItemsRemovedId = "request-property-prefix-items-removed"
)

func RequestPropertyPrefixItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PrefixItemsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "prefixItems")
			if len(info.schemaDiff.PrefixItemsDiff.Added) > 0 {
				result = append(result, info.newChange(RequestBodyPrefixItemsAddedId, []any{info.schemaDiff.PrefixItemsDiff.Added.String()}, "").
					WithSources(nil, revisionSource))
			}
			if len(info.schemaDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, info.newChange(RequestBodyPrefixItemsRemovedId, []any{info.schemaDiff.PrefixItemsDiff.Deleted.String()}, "").
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.PrefixItemsDiff == nil {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "prefixItems")

			if len(p.propertyDiff.PrefixItemsDiff.Added) > 0 {
				result = append(result, p.newChange(RequestPropertyPrefixItemsAddedId, []any{p.propertyDiff.PrefixItemsDiff.Added.String(), propName}, "").
					WithSources(nil, propRevisionSource))
			}
			if len(p.propertyDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, p.newChange(RequestPropertyPrefixItemsRemovedId, []any{p.propertyDiff.PrefixItemsDiff.Deleted.String(), propName}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
