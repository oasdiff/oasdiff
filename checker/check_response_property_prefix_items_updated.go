package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyPrefixItemsAddedId       = "response-body-prefix-items-added"
	ResponseBodyPrefixItemsRemovedId     = "response-body-prefix-items-removed"
	ResponsePropertyPrefixItemsAddedId   = "response-property-prefix-items-added"
	ResponsePropertyPrefixItemsRemovedId = "response-property-prefix-items-removed"
)

func ResponsePropertyPrefixItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if info.schemaDiff.PrefixItemsDiff != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "prefixItems")
			if len(info.schemaDiff.PrefixItemsDiff.Added) > 0 {
				result = append(result, info.newChange(ResponseBodyPrefixItemsAddedId, []any{info.schemaDiff.PrefixItemsDiff.Added.String(), info.responseStatus}, "").
					WithSources(nil, revisionSource))
			}
			if len(info.schemaDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, info.newChange(ResponseBodyPrefixItemsRemovedId, []any{info.schemaDiff.PrefixItemsDiff.Deleted.String(), info.responseStatus}, "").
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
				result = append(result, p.newChange(ResponsePropertyPrefixItemsAddedId, []any{p.propertyDiff.PrefixItemsDiff.Added.String(), propName, info.responseStatus}, "").
					WithSources(nil, propRevisionSource))
			}
			if len(p.propertyDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, p.newChange(ResponsePropertyPrefixItemsRemovedId, []any{p.propertyDiff.PrefixItemsDiff.Deleted.String(), propName, info.responseStatus}, "").
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
