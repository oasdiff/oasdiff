package checker

import (
	"github.com/oasdiff/oasdiff/diff"
	"slices"
)

const (
	RequestParameterXExtensibleEnumValueRemovedId = "request-parameter-x-extensible-enum-value-removed"
)

func RequestParameterXExtensibleEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
					if paramItem.SchemaDiff.ExtensionsDiff == nil {
						continue
					}
					if paramItem.SchemaDiff.ExtensionsDiff.Modified == nil {
						continue
					}
					if paramItem.SchemaDiff.ExtensionsDiff.Modified[diff.XExtensibleEnumExtension] == nil {
						continue
					}
					from, ok := paramItem.SchemaDiff.Base.Extensions[diff.XExtensibleEnumExtension].([]any)
					if !ok {
						continue
					}
					to, ok := paramItem.SchemaDiff.Revision.Extensions[diff.XExtensibleEnumExtension].([]any)
					if !ok {
						continue
					}

					fromSlice := make([]string, len(from))
					for i, item := range from {
						fromSlice[i] = item.(string)
					}

					toSlice := make([]string, len(to))
					for i, item := range to {
						toSlice[i] = item.(string)
					}

					deletedVals := make([]string, 0)
					for _, fromVal := range fromSlice {
						if !slices.Contains(toSlice, fromVal) {
							deletedVals = append(deletedVals, fromVal)
						}
					}

					for _, enumVal := range deletedVals {
						result = append(result, NewApiChange(
							RequestParameterXExtensibleEnumValueRemovedId,
							config,
							[]any{enumVal, paramLocation, paramName},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
					}
				}
			}
		}
	}
	return result
}
