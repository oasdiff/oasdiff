package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestHeaderPropertyBecameRequiredId = "request-header-property-became-required"
)

func RequestHeaderPropertyBecameRequiredCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.location != "header" {
			return
		}
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		if p.paramDiff.SchemaDiff.RequiredDiff != nil {
			for _, changedRequiredPropertyName := range p.paramDiff.SchemaDiff.RequiredDiff.Added {
				if p.paramDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
					continue
				}
				if p.paramDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName].Value.ReadOnly {
					continue
				}

				if p.paramDiff.SchemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
					// new added required properties processed via the new-required-request-header-property check
					continue
				}

				baseSource, revisionSource := SchemaAddedItemSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "required", changedRequiredPropertyName)
				result = append(result, p.opInfo.NewApiChange(
					RequestHeaderPropertyBecameRequiredId,
					[]any{p.name, changedRequiredPropertyName},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		checkModifiedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				requiredDiff := propertyDiff.RequiredDiff
				if requiredDiff == nil {
					return
				}
				for _, changedRequiredPropertyName := range requiredDiff.Added {
					if propertyDiff.Revision.Properties[changedRequiredPropertyName] == nil {
						continue
					}
					if propertyDiff.Revision.Properties[changedRequiredPropertyName].Value.ReadOnly {
						continue
					}
					propBaseSource, propRevisionSource := SchemaAddedItemSources(operationsSources, p.opInfo.methodDiff, propertyDiff, "required", changedRequiredPropertyName)
					result = append(result, p.opInfo.NewApiChange(
						RequestHeaderPropertyBecameRequiredId,
						[]any{p.name, propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				}
			})
	})
	return result
}
