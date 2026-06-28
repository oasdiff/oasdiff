package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterEnumValueAddedId           = "request-parameter-enum-value-added"
	RequestParameterEnumValueRemovedId         = "request-parameter-enum-value-removed"
	RequestParameterPropertyEnumValueAddedId   = "request-parameter-property-enum-value-added"
	RequestParameterPropertyEnumValueRemovedId = "request-parameter-property-enum-value-removed"
)

func RequestParameterEnumValueUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

					result = append(result, checkParameterEnumDiff(
						opInfo,
						paramItem.SchemaDiff.EnumDiff,
						paramItem.SchemaDiff,
						RequestParameterEnumValueRemovedId,
						RequestParameterEnumValueAddedId,
						func(enumVal any) []any { return []any{enumVal, paramLocation, paramName} },
					)...)

					checkModifiedPropertiesDiff(
						paramItem.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							result = append(result, checkParameterEnumDiff(
								opInfo,
								propertyDiff.EnumDiff,
								propertyDiff,
								RequestParameterPropertyEnumValueRemovedId,
								RequestParameterPropertyEnumValueAddedId,
								func(enumVal any) []any {
									return []any{enumVal, propertyFullName(propertyPath, propertyName), paramLocation, paramName}
								},
							)...)
						})
				}
			}
		}
	}
	return result
}

func checkParameterEnumDiff(
	opInfo opInfo,
	enumDiff *diff.EnumDiff,
	schemaDiff *diff.SchemaDiff,
	removedId, addedId string,
	makeArgs func(enumVal any) []any,
) Changes {
	result := make(Changes, 0)
	if enumDiff == nil {
		return result
	}

	for _, enumVal := range enumDiff.Deleted {
		baseSource, revisionSource := SchemaDeletedItemSources(opInfo.operationsSources, opInfo.methodDiff, schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
		result = append(result, NewApiChange(
			removedId,
			opInfo.config,
			makeArgs(enumVal),
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, revisionSource))
	}

	for _, enumVal := range enumDiff.Added {
		baseSource, revisionSource := SchemaAddedItemSources(opInfo.operationsSources, opInfo.methodDiff, schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
		result = append(result, NewApiChange(
			addedId,
			opInfo.config,
			makeArgs(enumVal),
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, revisionSource))
	}

	return result
}
