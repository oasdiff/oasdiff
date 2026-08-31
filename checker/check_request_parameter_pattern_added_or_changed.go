package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterPatternAddedId       = "request-parameter-pattern-added"
	RequestParameterPatternRemovedId     = "request-parameter-pattern-removed"
	RequestParameterPatternChangedId     = "request-parameter-pattern-changed"
	RequestParameterPatternGeneralizedId = "request-parameter-pattern-generalized"
	PatternChangedCommentId              = "pattern-changed-warn-comment"
	PatternAddedCommentId                = "pattern-added-error-comment"
)

func RequestParameterPatternAddedOrChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.SchemaDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "pattern")
		patternDiff := p.paramDiff.SchemaDiff.PatternDiff
		if patternDiff == nil {
			return
		}

		if patternDiff.From == "" {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterPatternAddedId,
				[]any{patternDiff.To, p.location, p.name},
				PatternAddedCommentId,
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(nil, revisionSource))
		} else if patternDiff.To == "" {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterPatternRemovedId,
				[]any{patternDiff.From, p.location, p.name},
				"",
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, nil))
		} else {
			id := RequestParameterPatternChangedId
			comment := PatternChangedCommentId

			if patternDiff.To == ".*" {
				id = RequestParameterPatternGeneralizedId
				comment = ""
			}

			result = append(result, p.opInfo.NewApiChange(
				id,
				[]any{p.location, p.name, patternDiff.From, patternDiff.To},
				comment,
			).WithSchema(p.paramDiff.SchemaDiff).WithSources(baseSource, revisionSource))
		}
	})
	return result
}
