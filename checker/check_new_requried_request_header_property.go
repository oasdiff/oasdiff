package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	NewRequiredRequestHeaderPropertyId = "new-required-request-header-property"
)

func NewRequiredRequestHeaderPropertyCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.location != "header" {
			return
		}
		baseSource, revisionSource := ParameterSources(operationsSources, p.opInfo.methodDiff, p.paramDiff)
		checkAddedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, newPropertyName string, newProperty *openapi3.Schema, parent *diff.SchemaDiff) {
				if newProperty.ReadOnly {
					return
				}
				if !slices.Contains(parent.Revision.Required, newPropertyName) {
					return
				}

				result = append(result, p.opInfo.NewApiChange(
					NewRequiredRequestHeaderPropertyId,
					[]any{p.name, propertyFullName(propertyPath, newPropertyName)},
					"",
				).WithSources(baseSource, revisionSource))
			})
	})
	return result
}
