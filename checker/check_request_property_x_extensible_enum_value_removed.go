package checker

import (
	"github.com/oasdiff/oasdiff/diff"
	"golang.org/x/exp/slices"
)

const (
	RequestPropertyXExtensibleEnumValueRemovedId = "request-property-x-extensible-enum-value-removed"
)

func RequestPropertyXExtensibleEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for _, mediaTypeDiff := range modifiedMediaTypes {
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						if propertyDiff.ExtensionsDiff == nil {
							return
						}
						if propertyDiff.ExtensionsDiff.Modified == nil {
							return
						}
						if propertyDiff.ExtensionsDiff.Modified[diff.XExtensibleEnumExtension] == nil {
							return
						}
						from, ok := propertyDiff.Base.Extensions[diff.XExtensibleEnumExtension].([]interface{})
						if !ok {
							return
						}
						to, ok := propertyDiff.Revision.Extensions[diff.XExtensibleEnumExtension].([]interface{})
						if !ok {
							return
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

						if propertyDiff.Revision.ReadOnly {
							return
						}
						for _, enumVal := range deletedVals {
							result = append(result, NewApiChange(
								RequestPropertyXExtensibleEnumValueRemovedId,
								config,
								[]any{enumVal, propertyFullName(propertyPath, propertyName)},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource))
						}
					})
			}
		}
	}
	return result
}
