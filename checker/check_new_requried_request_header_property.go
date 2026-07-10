package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	NewRequiredRequestHeaderPropertyId = "new-required-request-header-property"
)

func NewRequiredRequestHeaderPropertyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					baseSource, revisionSource := ParameterSources(operationsSources, operationItem, paramDiff)
					checkAddedPropertiesDiff(
						paramDiff.SchemaDiff,
						func(propertyPath string, newPropertyName string, newProperty *openapi3.Schema, parent *diff.SchemaDiff) {
							if newProperty.ReadOnly {
								return
							}
							if !slices.Contains(parent.Revision.Required, newPropertyName) {
								return
							}

							result = append(result, opInfo.NewApiChange(
								NewRequiredRequestHeaderPropertyId,
								[]any{paramName, propertyFullName(propertyPath, newPropertyName)},
								"",
							).WithSources(baseSource, revisionSource))
						})
				}
			}
		}
	}
	return result
}
