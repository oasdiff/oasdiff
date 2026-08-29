package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseBodyMultipleOfUnsetId           = "response-body-multiple-of-unset"
	ResponsePropertyMultipleOfUnsetId       = "response-property-multiple-of-unset"
	ResponseBodyMultipleOfChangedId         = "response-body-multiple-of-changed"
	ResponsePropertyMultipleOfChangedId     = "response-property-multiple-of-changed"
	ResponseBodyMultipleOfSpecializedId     = "response-body-multiple-of-specialized"
	ResponsePropertyMultipleOfSpecializedId = "response-property-multiple-of-specialized"
)

func ResponsePropertyMultipleOfUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		if multipleOfDiff := info.schemaDiff.MultipleOfDiff; multipleOfDiff != nil && multipleOfDiff.From != nil {
			baseSource, revisionSource := SchemaFieldSources(operationsSources, info.operationItem, info.schemaDiff, "multipleOf")
			switch {
			case multipleOfDiff.To == nil:
				result = append(result, info.newChange(
					ResponseBodyMultipleOfUnsetId,
					[]any{multipleOfDiff.From},
					"",
				).WithSources(baseSource, revisionSource))
			case isIntegerMultiple(multipleOfDiff.To.(float64), multipleOfDiff.From.(float64)):
				result = append(result, info.newChange(
					ResponseBodyMultipleOfSpecializedId,
					[]any{multipleOfDiff.From, multipleOfDiff.To},
					commentId(ResponseBodyMultipleOfSpecializedId),
				).WithSources(baseSource, revisionSource))
			default:
				result = append(result, info.newChange(
					ResponseBodyMultipleOfChangedId,
					[]any{multipleOfDiff.From, multipleOfDiff.To},
					"",
				).WithSources(baseSource, revisionSource))
			}
		}

		info.walkProperties(func(p propertyInfo) {
			multipleOfDiff := p.propertyDiff.MultipleOfDiff
			if multipleOfDiff == nil || multipleOfDiff.From == nil {
				return
			}
			if p.propertyDiff.Revision.WriteOnly {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "multipleOf")
			switch {
			case multipleOfDiff.To == nil:
				result = append(result, p.newChange(
					ResponsePropertyMultipleOfUnsetId,
					[]any{propName, multipleOfDiff.From, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			case isIntegerMultiple(multipleOfDiff.To.(float64), multipleOfDiff.From.(float64)):
				result = append(result, p.newChange(
					ResponsePropertyMultipleOfSpecializedId,
					[]any{propName, multipleOfDiff.From, multipleOfDiff.To, info.responseStatus},
					commentId(ResponsePropertyMultipleOfSpecializedId),
				).WithSources(propBaseSource, propRevisionSource))
			default:
				result = append(result, p.newChange(
					ResponsePropertyMultipleOfChangedId,
					[]any{propName, multipleOfDiff.From, multipleOfDiff.To, info.responseStatus},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
