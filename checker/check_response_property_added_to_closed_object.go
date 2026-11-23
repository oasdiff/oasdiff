package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyAddedToClosedObjectId = "response-property-added-to-closed-object"
)

// ResponsePropertyAddedToClosedObjectCheck reports a property added to a
// response object that was closed with additionalProperties: false. Clients
// that validate against the previous schema reject the new key, so unlike an
// addition to an open object this is breaking.
func ResponsePropertyAddedToClosedObjectCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		schemawalk.AddedProperties(
			info.schemaDiff,
			func(propertyPath string, propertyName string, propertyItem *openapi3.Schema, parent *diff.SchemaDiff, underAllOf bool) {
				if getAdditionalPropertiesMode(parent.Base) != additionalPropertiesNone {
					return
				}
				if propertyItem.WriteOnly {
					return
				}
				propName := schemawalk.PropertyFullName(propertyPath, propertyName)
				revisionSource := location.PropertySource(operationsSources, info.operationItem.Revision, propertyItem)
				result = append(result, info.newChange(
					ResponsePropertyAddedToClosedObjectId,
					[]any{propName, info.responseStatus},
					"",
				).WithSchema(info.schemaDiff, parent, propName).WithAllOfDisclaimers(underAllOf, nil).WithSources(nil, revisionSource))
			})
	})

	return result
}
