package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyRemovedId                = "request-property-removed"
	NewRequiredRequestPropertyId            = "new-required-request-property"
	NewRequiredRequestPropertyWithDefaultId = "new-required-request-property-with-default"
	NewOptionalRequestPropertyId            = "new-optional-request-property"
	RequestBodyWrappedInOneOfId             = "request-body-wrapped-in-one-of"

	RequestBodyWrappedInOneOfOriginalPreservedId = "request-body-wrapped-in-one-of-original-preserved"

	// Shared by the request and response wrapped-in-one-of-original-preserved
	// checks. The spec does not say whether the alternatives overlap, and
	// oneOf rejects a payload matching more than one, so a payload that was
	// valid before may or may not still be accepted.
	WrappedInOneOfOriginalPreservedCommentId = "wrapped-in-one-of-original-preserved-warning-comment"

	// Shared by the two "required request property with a default" checks. A
	// request that omits a required property is invalid under the new contract
	// whether or not the property has a default: the default is a server-side
	// fallback, not a rule that makes the omitted property valid. So these are
	// breaking (ERR), same as the no-default siblings, with this comment
	// explaining why the default does not make them safe.
	RequiredRequestPropertyWithDefaultCommentId = "required-request-property-with-default-error-comment"
)

func RequestPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// One finding per wrapped body, not per property. Which one depends on
		// whether an alternative still accepts everything the base did: if none
		// does, a previously valid payload matches nothing; if one does, the
		// payload matches it but oneOf rejects it should it match a second
		// alternative too, which the spec does not say.
		// A nullable wrapping also matches this shape, but the claim filter
		// drops this change there (KindStructure is claimed and became-nullable
		// reports it), so no explicit precedence check is needed.
		if w := info.schemaDiff.OneOfWrappingDiff; !w.Empty() {
			id, comment := RequestBodyWrappedInOneOfId, ""
			if w.OriginalPreserved {
				id, comment = RequestBodyWrappedInOneOfOriginalPreservedId, WrappedInOneOfOriginalPreservedCommentId
			}
			result = append(result, info.newChange(id, nil, comment).WithSources(nil, nil))
		}

		// checkDeletedPropertiesDiff / checkAddedPropertiesDiff handle the
		// added/removed property sides — different from info.walkProperties,
		// which delegates to checkModifiedPropertiesDiff. Used directly here.
		checkDeletedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				if propertyItem.ReadOnly {
					return
				}

				// A property that moved into a oneOf wrapping (#702) was not
				// removed from the contract, so the raw "property removed" finding
				// is a false positive. Suppress it; the breaking nature of the
				// wrapping is reported once per body as request-body-wrapped-in-one-of.
				if w := parent.OneOfWrappingDiff; w != nil && slices.Contains(w.MovedProperties, propertyName) {
					return
				}

				baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
				result = append(result, info.newChange(
					RequestPropertyRemovedId,
					[]any{propertyFullName(propertyPath, propertyName)},
					"",
				).WithSchema(parent).WithSources(baseSource, nil))
			})

		checkAddedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				if propertyItem.ReadOnly {
					return
				}

				propName := propertyFullName(propertyPath, propertyName)
				revisionSource := propertySource(operationsSources, info.operationItem.Revision, propertyItem)

				if slices.Contains(parent.Revision.Required, propertyName) {
					if propertyItem.Default == nil {
						result = append(result, info.newChange(
							NewRequiredRequestPropertyId,
							[]any{propName},
							"",
						).WithSchema(parent).WithSources(nil, revisionSource))
					} else {
						result = append(result, info.newChange(
							NewRequiredRequestPropertyWithDefaultId,
							[]any{propName},
							RequiredRequestPropertyWithDefaultCommentId,
						).WithSchema(parent).WithSources(nil, revisionSource))
					}
				} else {
					result = append(result, info.newChange(
						NewOptionalRequestPropertyId,
						[]any{propName},
						"",
					).WithSchema(parent).WithSources(nil, revisionSource))
				}
			})
	})

	return result
}
