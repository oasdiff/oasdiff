package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseOptionalPropertyRemovedId          = "response-optional-property-removed"
	ResponseOptionalWriteOnlyPropertyRemovedId = "response-optional-write-only-property-removed"
	ResponseOptionalPropertyAddedId            = "response-optional-property-added"
	ResponseOptionalWriteOnlyPropertyAddedId   = "response-optional-write-only-property-added"
)

func ResponseOptionalPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// checkDeletedPropertiesDiff / checkAddedPropertiesDiff handle the
		// added/removed property sides; info.walkProperties only delegates
		// to checkModifiedPropertiesDiff so the primitives are called
		// directly inside the walker callback.
		checkDeletedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				if slices.Contains(parent.Base.Required, propertyName) {
					// covered by response-required-property-removed
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
				id := ResponseOptionalPropertyRemovedId
				if propertyItem.WriteOnly {
					id = ResponseOptionalWriteOnlyPropertyRemovedId
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
				if slices.Contains(parent.Revision.Required, propertyName) {
					// covered by response-required-property-added
					return
				}
				id := ResponseOptionalPropertyAddedId
				if propertyItem.WriteOnly {
					id = ResponseOptionalWriteOnlyPropertyAddedId
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
