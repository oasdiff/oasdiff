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
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						patternDiff := propertyDiff.PatternDiff
						if patternDiff == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						if patternDiff.To == "" {
							result = append(result, NewApiChange(
								RequestPropertyPatternRemovedId,
								config,
								[]any{patternDiff.From, propName},
								"",
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						} else if patternDiff.From == "" {
							result = append(result, NewApiChange(
								RequestPropertyPatternAddedId,
								config,
								[]any{patternDiff.To, propName},
								PatternAddedCommentId,
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						} else {

							id := RequestPropertyPatternChangedId
							comment := PatternChangedCommentId

							if patternDiff.To == ".*" {
								id = RequestPropertyPatternGeneralizedId
								comment = ""
							}

							result = append(result, NewApiChange(
								id,
								config,
								[]any{propName, patternDiff.From, patternDiff.To},
								comment,
								operationsSources,
								operationItem.Revision,
								operation,
								path,
							).WithSources(baseSource, revisionSource).WithDetails(mediaTypeDetails))
						}
					})
			}
		}
	}
	return result
}
