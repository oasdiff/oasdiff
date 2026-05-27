package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyPatternRemovedId     = "request-property-pattern-removed"
	RequestPropertyPatternAddedId       = "request-property-pattern-added"
	RequestPropertyPatternChangedId     = "request-property-pattern-changed"
	RequestPropertyPatternGeneralizedId = "request-property-pattern-generalized"
)

func RequestPropertyPatternUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			patternDiff := p.propertyDiff.PatternDiff
			if patternDiff == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "pattern")

			if patternDiff.To == "" {
				result = append(result, p.newChange(
					RequestPropertyPatternRemovedId,
					[]any{patternDiff.From, propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
			} else if patternDiff.From == "" {
				result = append(result, p.newChange(
					RequestPropertyPatternAddedId,
					[]any{patternDiff.To, propName},
					PatternAddedCommentId,
				).WithSources(propBaseSource, propRevisionSource))
			} else {
				id := RequestPropertyPatternChangedId
				comment := PatternChangedCommentId
				if patternDiff.To == ".*" {
					id = RequestPropertyPatternGeneralizedId
					comment = ""
				}
				result = append(result, p.newChange(
					id,
					[]any{propName, patternDiff.From, patternDiff.To},
					comment,
				).WithSources(propBaseSource, propRevisionSource))
			}
		})
	})

	return result
}
