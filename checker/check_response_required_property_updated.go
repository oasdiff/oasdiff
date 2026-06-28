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
	ResponseBodyWrappedInOneOfId               = "response-body-wrapped-in-one-of"
)

func ResponseRequiredPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// Wrapping a concrete object response body into a oneOf (#702) is a
		// breaking restructuring: a field that was previously guaranteed in the
		// response may now be absent depending on which alternative matches.
		// Emit one breaking finding per wrapped body (not per property). This
		// is the single emission point for the response side, mirroring the
		// request side which emits in RequestPropertyUpdatedCheck.
		if !info.schemaDiff.OneOfWrappingDiff.Empty() {
			result = append(result, info.newChange(
				ResponseBodyWrappedInOneOfId,
				nil,
				"",
			).WithSources(nil, nil))
		}

		// checkDeletedPropertiesDiff / checkAddedPropertiesDiff walk
		// properties that were dropped or introduced entirely, not just
		// modified ones — different from info.walkProperties, which
		// delegates to checkModifiedPropertiesDiff. Used directly here.
		checkDeletedPropertiesDiff(
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

				// A property that moved into a oneOf wrapping (#702) was not
				// removed from the contract, so the raw "property removed"
				// finding is a false positive. Suppress it; the breaking nature
				// of the wrapping is reported once per body as
				// response-body-wrapped-in-one-of.
				if w := parent.OneOfWrappingDiff; w != nil && slices.Contains(w.MovedProperties, propertyName) {
					return
				}

				baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
				result = append(result, info.newChange(
					id,
					[]any{propertyFullName(propertyPath, propertyName), info.responseStatus},
					"",
				).WithSources(baseSource, nil))
			})
		checkAddedPropertiesDiff(
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
