package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterEnumValueAddedId   = "request-parameter-enum-value-added"
	RequestParameterEnumValueRemovedId = "request-parameter-enum-value-removed"
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
			for paramLocation, paramItems := range operationItem.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {
					if paramItem.SchemaDiff == nil {
						continue
					}
					enumDiff := paramItem.SchemaDiff.EnumDiff
					if enumDiff == nil {
						continue
					}
					for _, enumVal := range enumDiff.Deleted {
						baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, operationItem, paramItem.SchemaDiff, "enum", fmt.Sprintf("%v", enumVal))
						result = append(result, NewApiChange(
							RequestParameterEnumValueRemovedId,
							config,
							[]any{enumVal, paramLocation, paramName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}
					for _, enumVal := range enumDiff.Added {
						baseSource, revisionSource := SchemaAddedItemSources(operationsSources, operationItem, paramItem.SchemaDiff, "enum", fmt.Sprintf("%v", enumVal))
						result = append(result, NewApiChange(
							RequestParameterEnumValueAddedId,
							config,
							[]any{enumVal, paramLocation, paramName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, revisionSource))
					}
				}
			}
		}
	}
	return result
}
