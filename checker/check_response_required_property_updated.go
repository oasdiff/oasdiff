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

	ResponseBodyWrappedInOneOfOriginalPreservedId = "response-body-wrapped-in-one-of-original-preserved"
)

func ResponseRequiredPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// One finding per wrapped body, not per property, and the single
		// emission point for the response side. Which one depends on whether an
		// alternative still describes everything the base did: if none does, a
		// field that was guaranteed may now be absent; if one does, the
		// response still matches it, and the spec does not say whether it also
		// matches a second alternative, which oneOf rejects.
		// A nullable wrapping also matches this shape, but the claim filter
		// drops this change there (KindStructure is claimed and became-nullable
		// reports it), so no explicit precedence check is needed.
		if w := info.schemaDiff.OneOfWrappingDiff; !w.Empty() {
			id, comment := ResponseBodyWrappedInOneOfId, ""
			if w.OriginalPreserved {
				id, comment = ResponseBodyWrappedInOneOfOriginalPreservedId, WrappedInOneOfOriginalPreservedCommentId
			}
			result = append(result, info.newChange(id, nil, comment).WithSources(nil, nil))
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
				).WithSchema(parent).WithSources(baseSource, nil))
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
				).WithSchema(parent).WithSources(nil, revisionSource))
			})
	})

	return result
}
