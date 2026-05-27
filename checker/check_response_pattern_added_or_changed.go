package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyPatternAddedId   = "response-property-pattern-added"
	ResponsePropertyPatternChangedId = "response-property-pattern-changed"
	ResponsePropertyPatternRemovedId = "response-property-pattern-removed"
)

func ResponsePatternAddedOrChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		info.walkProperties(func(p propertyInfo) {
			patternDiff := p.propertyDiff.PatternDiff
			if patternDiff == nil {
				return
			}

			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "pattern")
			propName := propertyFullName(p.propertyPath, p.propertyName)

			id := ResponsePropertyPatternChangedId
			args := []any{propName, patternDiff.From, patternDiff.To, info.responseStatus}
			if patternDiff.To == "" || patternDiff.To == nil {
				id = ResponsePropertyPatternRemovedId
				args = []any{propName, patternDiff.From, info.responseStatus}
			} else if patternDiff.From == "" || patternDiff.From == nil {
				id = ResponsePropertyPatternAddedId
				args = []any{propName, patternDiff.To, info.responseStatus}
			}

			result = append(result, p.newChange(id, args, "").
				WithSources(propBaseSource, propRevisionSource))
		})
	})

	return result
}
