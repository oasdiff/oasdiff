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
)

func RequestPropertyUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// CheckDeletedPropertiesDiff / CheckAddedPropertiesDiff handle the
		// added/removed property sides — different from info.walkProperties,
		// which delegates to CheckModifiedPropertiesDiff. Used directly here.
		CheckDeletedPropertiesDiff(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff) {
				if propertyItem.ReadOnly {
					return
				}

				// A property that moved into a oneOf wrapping (#702) was not
				// removed from the contract, so the raw "property removed" finding
				// is a false positive. Suppress it; if the property was required
				// and is no longer required in every alternative, report that it
				// became optional instead.
				if w := parent.OneOfWrappingDiff; w != nil && slices.Contains(w.MovedProperties, propertyName) {
					if slices.Contains(w.RequiredBecameOptional, propertyName) {
						baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
						result = append(result, info.newChange(
							RequestPropertyBecameOptionalId,
							[]any{propertyFullName(propertyPath, propertyName)},
							"",
						).WithSources(baseSource, nil))
					}
					return
				}

				baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
				result = append(result, info.newChange(
					RequestPropertyRemovedId,
					[]any{propertyFullName(propertyPath, propertyName)},
					"",
				).WithSources(baseSource, nil))
			})

		CheckAddedPropertiesDiff(
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
						).WithSources(nil, revisionSource))
					} else {
						result = append(result, info.newChange(
							NewRequiredRequestPropertyWithDefaultId,
							[]any{propName},
							"",
						).WithSources(nil, revisionSource))
					}
				} else {
					result = append(result, info.newChange(
						NewOptionalRequestPropertyId,
						[]any{propName},
						"",
					).WithSources(nil, revisionSource))
				}
			})
	})

	return result
}
