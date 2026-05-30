package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameOptionalId          = "response-property-became-optional"
	ResponseWriteOnlyPropertyBecameOptionalId = "response-write-only-property-became-optional"
)

func ResponsePropertyBecameOptionalCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if requiredDiff := info.schemaDiff.RequiredDiff; requiredDiff != nil {
			for _, changedRequiredPropertyName := range requiredDiff.Deleted {
				id := ResponsePropertyBecameOptionalId
				if info.schemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
					// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
					continue
				}
				if info.schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.WriteOnly {
					id = ResponseWriteOnlyPropertyBecameOptionalId
				}

				baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, info.operationItem, info.schemaDiff, "required", changedRequiredPropertyName)
				result = append(result, info.newChange(
					id,
					[]any{changedRequiredPropertyName, info.responseStatus},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			requiredDiff := p.propertyDiff.RequiredDiff
			if requiredDiff == nil {
				return
			}
			for _, changedRequiredPropertyName := range requiredDiff.Deleted {

				if p.propertyDiff.Base.Properties[changedRequiredPropertyName] == nil {
					continue
				}

				if p.propertyDiff.Revision.Properties[changedRequiredPropertyName] == nil {
					// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
					continue
				}

				id := ResponsePropertyBecameOptionalId

				if p.propertyDiff.Base.Properties[changedRequiredPropertyName].Value.WriteOnly {
					id = ResponseWriteOnlyPropertyBecameOptionalId
				}

				propBaseSource, propRevisionSource := SchemaDeletedItemSources(operationsSources, info.operationItem, p.propertyDiff, "required", changedRequiredPropertyName)
				result = append(result, p.newChange(
					id,
					[]any{propertyFullName(p.propertyPath, propertyFullName(p.propertyName, changedRequiredPropertyName)), info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
