package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyBecameRequiredId            = "request-property-became-required"
	RequestPropertyBecameRequiredWithDefaultId = "request-property-became-required-with-default"
	RequestPropertyBecameOptionalId            = "request-property-became-optional"
)

func RequestPropertyRequiredUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		processRequiredDiff := func(schemaDiff *diff.SchemaDiff, propertyPath string, propertyName string) {
			if schemaDiff.RequiredDiff == nil {
				return
			}
			for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Added {
				if !changedRequiredPropertyRelevant(schemaDiff, changedRequiredPropertyName) {
					continue
				}
				srcBase, srcRevision := SchemaAddedItemSources(operationsSources, info.operationItem, schemaDiff, "required", changedRequiredPropertyName)
				args := []any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))}
				if schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.Default == nil {
					result = append(result, info.newChange(RequestPropertyBecameRequiredId, args, "").
						WithSources(srcBase, srcRevision))
				} else {
					// The property has a default value, but a request that omits a
					// required property is still invalid under the new contract (the
					// default is a server-side fallback, not a validity rule). So this
					// is breaking, same as the no-default case, with a comment
					// explaining why the default does not make it safe.
					result = append(result, info.newChange(RequestPropertyBecameRequiredWithDefaultId, args, RequiredRequestPropertyWithDefaultCommentId).
						WithSources(srcBase, srcRevision))
				}
			}
			for _, changedRequiredPropertyName := range schemaDiff.RequiredDiff.Deleted {
				if !changedRequiredPropertyRelevant(schemaDiff, changedRequiredPropertyName) {
					continue
				}
				srcBase, srcRevision := SchemaDeletedItemSources(operationsSources, info.operationItem, schemaDiff, "required", changedRequiredPropertyName)
				args := []any{propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))}
				result = append(result, info.newChange(RequestPropertyBecameOptionalId, args, "").
					WithSources(srcBase, srcRevision))
			}
		}

		processRequiredDiff(info.schemaDiff, "", "")

		info.walkProperties(func(p propertyInfo) {
			processRequiredDiff(p.propertyDiff, p.propertyPath, p.propertyName)
		})
	})

	return result
}

func changedRequiredPropertyRelevant(schemaDiff *diff.SchemaDiff, changedRequiredPropertyName string) bool {
	if schemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
		// it is a new property, checked by the new-required-request-property check
		return false
	}
	if schemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
		// property was removed, checked by request-property-removed
		return false
	}
	if schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.ReadOnly {
		// property is read-only, not relevant in requests
		return false
	}

	return true
}
