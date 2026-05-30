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
				if !propertyItem.ReadOnly {
					baseSource := propertySource(operationsSources, info.operationItem.Base, propertyItem)
					result = append(result, info.newChange(
						RequestPropertyRemovedId,
						[]any{propertyFullName(propertyPath, propertyName)},
						"",
					).WithSources(baseSource, nil))
				}
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
