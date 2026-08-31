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
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}

		result = append(result, checkParameterEnumDiff(
			p.opInfo,
			p.paramDiff.SchemaDiff.EnumDiff,
			p.paramDiff.SchemaDiff,
			RequestParameterEnumValueRemovedId,
			RequestParameterEnumValueAddedId,
			func(enumVal any) []any { return []any{enumVal, p.location, p.name} },
		)...)

		checkModifiedPropertiesDiff(
			p.paramDiff.SchemaDiff,
			func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
				result = append(result, checkParameterEnumDiff(
					p.opInfo,
					propertyDiff.EnumDiff,
					propertyDiff,
					RequestParameterPropertyEnumValueRemovedId,
					RequestParameterPropertyEnumValueAddedId,
					func(enumVal any) []any {
						return []any{enumVal, propertyFullName(propertyPath, propertyName), p.location, p.name}
					},
				)...)
			})
	})
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
		result = append(result, opInfo.NewApiChange(
			removedId,
			makeArgs(enumVal),
			"",
		).WithSchema(schemaDiff).WithSources(baseSource, revisionSource))
	}

	for _, enumVal := range enumDiff.Added {
		baseSource, revisionSource := SchemaAddedItemSources(opInfo.operationsSources, opInfo.methodDiff, schemaDiff, "enum", fmt.Sprintf("%v", enumVal))
		result = append(result, opInfo.NewApiChange(
			addedId,
			makeArgs(enumVal),
			"",
		).WithSchema(schemaDiff).WithSources(baseSource, revisionSource))
	}

	return result
}
