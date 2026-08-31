package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterListOfTypesWidenedId          = "request-parameter-list-of-types-widened"
	RequestParameterListOfTypesNarrowedId         = "request-parameter-list-of-types-narrowed"
	RequestParameterPropertyListOfTypesWidenedId  = "request-parameter-property-list-of-types-widened"
	RequestParameterPropertyListOfTypesNarrowedId = "request-parameter-property-list-of-types-narrowed"
)

func RequestParameterListOfTypesChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		param := p.paramDiff.Revision
		if param == nil {
			return
		}

		// Check parameter schema
		changes := checkParameterListOfTypesChange(
			p.opInfo,
			p.paramDiff,
			param,
		)
		result = append(result, changes...)

		// Check parameter properties
		if p.paramDiff.SchemaDiff != nil {
			checkModifiedPropertiesDiff(
				p.paramDiff.SchemaDiff,
				func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
					changes := checkParameterPropertyListOfTypesChange(
						p.opInfo,
						propertyPath,
						propertyName,
						propertyDiff,
						param,
					)
					result = append(result, changes...)
				})
		}
	})
	return result
}
