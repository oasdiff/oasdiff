package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterBecameNullableId            = "request-parameter-became-nullable"
	RequestParameterBecameNotNullableId         = "request-parameter-became-not-nullable"
	RequestParameterPropertyBecameNullableId    = "request-parameter-property-became-nullable"
	RequestParameterPropertyBecameNotNullableId = "request-parameter-property-became-not-nullable"
)

func RequestParameterBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			if operationItem.ParametersDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for paramLocation, paramItems := range operationItem.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {
					if paramItem.SchemaDiff == nil {
						continue
					}

					if id := nullabilityChangeId(paramItem.SchemaDiff, RequestParameterBecameNullableId, RequestParameterBecameNotNullableId); id != "" {
						baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, paramItem.SchemaDiff, "nullable")
						result = append(result, opInfo.NewApiChange(
							id,
							[]any{paramLocation, paramName},
							"",
						).WithSchema(paramItem.SchemaDiff).WithSources(baseSource, revisionSource))
					}

					checkModifiedPropertiesDiff(
						paramItem.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							if propertyDiff == nil || propertyDiff.Base == nil || propertyDiff.Revision == nil {
								return
							}
							if id := nullabilityChangeId(propertyDiff, RequestParameterPropertyBecameNullableId, RequestParameterPropertyBecameNotNullableId); id != "" {
								baseSource, revisionSource := SchemaFieldSources(operationsSources, operationItem, propertyDiff, "nullable")
								result = append(result, opInfo.NewApiChange(
									id,
									[]any{propertyFullName(propertyPath, propertyName), paramLocation, paramName},
									"",
								).WithSchema(propertyDiff).WithSources(baseSource, revisionSource))
							}
						})
				}
			}
		}
	}
	return result
}
