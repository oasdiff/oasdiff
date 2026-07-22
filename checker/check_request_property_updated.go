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

	// Shared by the two "required request property with a default" checks. A
	// required property with a default is a self-contradictory contract, and
	// whether omitting it breaks a client depends on the server (does it enforce
	// the field or apply the default?), which the spec does not say. So these are
	// warnings, not safe, with this comment explaining the condition.
	RequiredRequestPropertyWithDefaultCommentId = "required-request-property-with-default-warn-comment"
)

func RequestPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// Wrapping a concrete object request body into a oneOf (#702) is a
		// breaking restructuring: under oneOf a previously valid payload can
		// match multiple overlapping alternatives and be rejected. Emit one
		// breaking finding per wrapped body (not per property).
		// A nullable wrapping also matches this shape, but the claim filter
		// drops this change there (KindStructure is claimed and became-nullable
		// reports it), so no explicit precedence check is needed.
		if !info.schemaDiff.OneOfWrappingDiff.Empty() {
			result = append(result, info.newChange(
				RequestBodyWrappedInOneOfId,
				nil,
				"",
			).WithSources(nil, nil))
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
