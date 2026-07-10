package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestHeaderPropertyBecameRequiredId = "request-header-property-became-required"
)

func RequestHeaderPropertyBecameRequiredCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {

			if operationItem.ParametersDiff == nil {
				continue
			}

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)

			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {

				if paramLocation != "header" {
					continue
				}

				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}

					if paramDiff.SchemaDiff.RequiredDiff != nil {
						for _, changedRequiredPropertyName := range paramDiff.SchemaDiff.RequiredDiff.Added {
							if paramDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName] == nil {
								continue
							}
							if paramDiff.SchemaDiff.Revision.Properties[changedRequiredPropertyName].Value.ReadOnly {
								continue
							}

							if paramDiff.SchemaDiff.Base.Properties[changedRequiredPropertyName] == nil {
								// new added required properties processed via the new-required-request-header-property check
								continue
							}

							baseSource, revisionSource := SchemaAddedItemSources(operationsSources, operationItem, paramDiff.SchemaDiff, "required", changedRequiredPropertyName)
							result = append(result, opInfo.NewApiChange(
								RequestHeaderPropertyBecameRequiredId,
								[]any{paramName, changedRequiredPropertyName},
								"",
							).WithSources(baseSource, revisionSource))
						}
					}

					checkModifiedPropertiesDiff(
						paramDiff.SchemaDiff,
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
								propBaseSource, propRevisionSource := SchemaAddedItemSources(operationsSources, operationItem, propertyDiff, "required", changedRequiredPropertyName)
								result = append(result, opInfo.NewApiChange(
									RequestHeaderPropertyBecameRequiredId,
									[]any{paramName, propertyFullName(propertyPath, propertyFullName(propertyName, changedRequiredPropertyName))},
									"",
								).WithSources(propBaseSource, propRevisionSource))
							}
						})
				}
			}
		}
	}
	return result
}
