package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMinItemsUnsetId     = "response-body-min-items-unset"
	ResponsePropertyMinItemsUnsetId = "response-property-min-items-unset"
)

func ResponsePropertyMinItemsUnsetCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if minItemsDiff := info.schemaDiff.MinItemsDiff; minItemsDiff != nil &&
			minItemsDiff.From != nil && minItemsDiff.To == nil {
			baseSource, _ := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "minItems")
			result = append(result, info.newChange(
				ResponseBodyMinItemsUnsetId,
				[]any{minItemsDiff.From},
				"",
			).WithSources(baseSource, nil))
		}

		info.walkProperties(func(p propertyInfo) {
			minItemsDiff := p.propertyDiff.MinItemsDiff
			if minItemsDiff == nil || minItemsDiff.To != nil || minItemsDiff.From == nil {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propBaseSource, _ := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "minItems")
			result = append(result, p.newChange(
				ResponsePropertyMinItemsUnsetId,
				[]any{propertyFullName(p.propertyPath, p.propertyName), minItemsDiff.From, info.responseStatus},
				"",
			).WithSources(propBaseSource, nil))
		})
	})

	return result
}
