package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseRequiredPropertyRemovedId          = "response-required-property-removed"
	ResponseRequiredWriteOnlyPropertyRemovedId = "response-required-write-only-property-removed"
	ResponseRequiredPropertyAddedId            = "response-required-property-added"
	ResponseRequiredWriteOnlyPropertyAddedId   = "response-required-write-only-property-added"
)

func ResponseRequiredPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// CheckDeletedPropertiesDiff / CheckAddedPropertiesDiff walk
		// properties that were dropped or introduced entirely, not just
		// modified ones — different from info.walkProperties, which
		// delegates to CheckModifiedPropertiesDiff. Used directly here.
		CheckDeletedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				id := ResponseRequiredPropertyRemovedId
				if propertyItem.WriteOnly {
					id = ResponseRequiredWriteOnlyPropertyRemovedId
				}
				if !slices.Contains(parent.Base.Required, propertyName) {
					// Covered by response-optional-property-removed
					return
				}

				baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
				result = append(result, info.newChange(
					id,
					[]any{propertyFullName(propertyPath, propertyName), info.responseStatus},
					"",
				).WithSources(baseSource, nil))
			})
		CheckAddedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				id := ResponseRequiredPropertyAddedId
				if propertyItem.WriteOnly {
					id = ResponseRequiredWriteOnlyPropertyAddedId
				}
				if !slices.Contains(parent.Revision.Required, propertyName) {
					// Covered by response-optional-property-added
					return
				}

				revisionSource := propertySource(operationsSources, info.operationItem.Revision, propertyItem)
				result = append(result, info.newChange(
					id,
					[]any{propertyFullName(propertyPath, propertyName), info.responseStatus},
					"",
				).WithSources(nil, revisionSource))
			})
	})

	return result
}
