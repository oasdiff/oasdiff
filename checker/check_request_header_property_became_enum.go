package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestHeaderPropertyBecameEnumId = "request-header-property-became-enum"
)

func RequestHeaderPropertyBecameEnumCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			for paramLocation, paramDiffs := range operationItem.ParametersDiff.Modified {

				if paramLocation != "header" {
					continue
				}

				for paramName, paramDiff := range paramDiffs {
					if paramDiff.SchemaDiff == nil {
						continue
					}

					baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramDiff.SchemaDiff, "enum")
					if paramDiff.SchemaDiff.EnumDiff != nil && paramDiff.SchemaDiff.EnumDiff.EnumAdded {
						result = append(result, NewApiChange(
							RequestHeaderPropertyBecameEnumId,
							config,
							[]any{paramName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}

					CheckModifiedPropertiesDiff(
						paramDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {

							if enumDiff := propertyDiff.EnumDiff; enumDiff == nil || !enumDiff.EnumAdded {
								return
							}

							propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "enum")
							result = append(result, NewApiChange(
								RequestHeaderPropertyBecameEnumId,
								config,
								[]any{paramName, propertyFullName(propertyPath, propertyName)},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(propBaseSource, propRevisionSource))
						})
				}
			}
		}
	}
	return result
}
