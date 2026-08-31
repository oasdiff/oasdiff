package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterXExtensibleEnumValueRemovedId = "request-parameter-x-extensible-enum-value-removed"
)

func RequestParameterXExtensibleEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		if p.paramDiff.SchemaDiff.ExtensionsDiff == nil {
			return
		}
		if p.paramDiff.SchemaDiff.ExtensionsDiff.Modified == nil {
			return
		}
		if p.paramDiff.SchemaDiff.ExtensionsDiff.Modified[diff.XExtensibleEnumExtension] == nil {
			return
		}
		from, ok := p.paramDiff.SchemaDiff.Base.Extensions[diff.XExtensibleEnumExtension].([]any)
		if !ok {
			return
		}
		to, ok := p.paramDiff.SchemaDiff.Revision.Extensions[diff.XExtensibleEnumExtension].([]any)
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

		for _, enumVal := range deletedVals {
			baseSource, revisionSource := SchemaDeletedItemSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, diff.XExtensibleEnumExtension, enumVal)
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterXExtensibleEnumValueRemovedId,
				[]any{enumVal, p.location, p.name},
				"",
			).WithSources(baseSource, revisionSource))
		}
	})
	return result
}
