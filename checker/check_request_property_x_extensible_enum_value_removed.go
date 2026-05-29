package checker

import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyXExtensibleEnumValueRemovedId = "request-property-x-extensible-enum-value-removed"
)

func RequestPropertyXExtensibleEnumValueRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.ExtensionsDiff == nil {
				return
			}
			if p.propertyDiff.ExtensionsDiff.Modified == nil {
				return
			}
			if p.propertyDiff.ExtensionsDiff.Modified[diff.XExtensibleEnumExtension] == nil {
				return
			}
			from, ok := p.propertyDiff.Base.Extensions[diff.XExtensibleEnumExtension].([]any)
			if !ok {
				return
			}
			to, ok := p.propertyDiff.Revision.Extensions[diff.XExtensibleEnumExtension].([]any)
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

			if p.propertyDiff.Revision.ReadOnly {
				return
			}
			propBaseSource, propRevisionSource := SchemaSources(operationsSources, info.operationItem, p.propertyDiff)
			for _, enumVal := range deletedVals {
				result = append(result, p.newChange(
					RequestPropertyXExtensibleEnumValueRemovedId,
					[]any{enumVal, propertyFullName(p.propertyPath, p.propertyName)},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
