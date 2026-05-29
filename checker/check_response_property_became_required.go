package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyBecameRequiredId          = "response-property-became-required"
	ResponseWriteOnlyPropertyBecameRequiredId = "response-write-only-property-became-required"
)

func ResponsePropertyBecameRequiredCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if requiredDiff := info.schemaDiff.RequiredDiff; requiredDiff != nil {
			for _, changedRequiredPropertyName := range requiredDiff.Added {
				if info.schemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
					// added properties are processed by ResponseRequiredPropertyUpdatedCheck check
					continue
				}
				if info.schemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
					// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
					continue
				}
				id := ResponsePropertyBecameRequiredId
				if info.schemaDiff.Revision.Properties[changedRequiredPropertyName].Value.WriteOnly {
					id = ResponseWriteOnlyPropertyBecameRequiredId
				}

				baseSource, revisionSource := SchemaAddedItemSources(operationsSources, info.operationItem, info.schemaDiff, "required", changedRequiredPropertyName)
				result = append(result, info.newChange(
					id,
					[]any{propertyFullName("", changedRequiredPropertyName), info.responseStatus},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			requiredDiff := p.propertyDiff.RequiredDiff
			if requiredDiff == nil {
				return
			}
			for _, changedRequiredPropertyName := range requiredDiff.Added {
				if p.propertyDiff.Base.Properties[changedRequiredPropertyName] == nil {
					// added properties are processed by ResponseRequiredPropertyUpdatedCheck check
					continue
				}
				if p.propertyDiff.Revision.Properties[changedRequiredPropertyName] == nil {
					// removed properties processed by the ResponseRequiredPropertyUpdatedCheck check
					continue
				}
				id := ResponsePropertyBecameRequiredId
				if p.propertyDiff.Base.Properties[changedRequiredPropertyName].Value.WriteOnly {
					id = ResponseWriteOnlyPropertyBecameRequiredId
				}

				propBaseSource, propRevisionSource := SchemaAddedItemSources(operationsSources, info.operationItem, p.propertyDiff, "required", changedRequiredPropertyName)
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
