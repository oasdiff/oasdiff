package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyPrefixItemsAddedId       = "request-body-prefix-items-added"
	RequestBodyPrefixItemsRemovedId     = "request-body-prefix-items-removed"
	RequestPropertyPrefixItemsAddedId   = "request-property-prefix-items-added"
	RequestPropertyPrefixItemsRemovedId = "request-property-prefix-items-removed"

	// Shared by the request and response prefixItems checks. A position no
	// prefixItems entry covers is governed by items, so an added entry may
	// constrain a position that was open or open one that items closed, and
	// the two schemas cannot be ordered. The direction is undetermined.
	PrefixItemsChangedCommentId = "prefix-items-changed-warning-comment"
)

func RequestPropertyPrefixItemsUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if prefixItemsChangedContract(info.schemaDiff) {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "prefixItems")
			if len(info.schemaDiff.PrefixItemsDiff.Added) > 0 {
				result = append(result, info.newChange(RequestBodyPrefixItemsAddedId, []any{info.schemaDiff.PrefixItemsDiff.Added.String()}, PrefixItemsChangedCommentId).
					WithSources(nil, revisionSource))
			}
			if len(info.schemaDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, info.newChange(RequestBodyPrefixItemsRemovedId, []any{info.schemaDiff.PrefixItemsDiff.Deleted.String()}, PrefixItemsChangedCommentId).
					WithSources(baseSource, nil))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			if !prefixItemsChangedContract(p.propertyDiff) {
				return
			}
			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "prefixItems")

			if len(p.propertyDiff.PrefixItemsDiff.Added) > 0 {
				result = append(result, p.newChange(RequestPropertyPrefixItemsAddedId, []any{p.propertyDiff.PrefixItemsDiff.Added.String(), propName}, PrefixItemsChangedCommentId).
					WithSources(nil, propRevisionSource))
			}
			if len(p.propertyDiff.PrefixItemsDiff.Deleted) > 0 {
				result = append(result, p.newChange(RequestPropertyPrefixItemsRemovedId, []any{p.propertyDiff.PrefixItemsDiff.Deleted.String(), propName}, PrefixItemsChangedCommentId).
					WithSources(propBaseSource, nil))
			}
		})
	})

	return result
}
