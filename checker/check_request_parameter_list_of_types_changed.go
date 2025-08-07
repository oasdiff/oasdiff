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

			// Check modified parameters
			for _, paramDiffs := range operationItem.ParametersDiff.Modified {
				for _, paramDiff := range paramDiffs {
					param := paramDiff.Revision
					if param == nil {
						continue
					}

					// Check parameter schema
					changes := checkParameterListOfTypesChange(
						paramDiff,
						param,
						config,
						operationsSources,
						operationItem,
						operation,
						path,
					)
					result = append(result, changes...)

					// Check parameter properties
					if paramDiff.SchemaDiff != nil {
						CheckModifiedPropertiesDiff(
							paramDiff.SchemaDiff,
							func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
								changes := checkParameterPropertyListOfTypesChange(
									propertyPath,
									propertyName,
									propertyDiff,
									param,
									config,
									operationsSources,
									operationItem,
									operation,
									path,
								)
								result = append(result, changes...)
							})
					}
				}
			}
		}
	}
	return result
}
