package checker

import (
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeNameChangedId     = "response-media-type-name-changed"
	ResponseMediaTypeNameGeneralizedId = "response-media-type-name-generalized"
	ResponseMediaTypeNameSpecializedId = "response-media-type-name-specialized"

	ResponseMediaTypeParameterAddedId   = "response-media-type-parameter-added"
	ResponseMediaTypeParameterRemovedId = "response-media-type-parameter-removed"
	ResponseMediaTypeParameterChangedId = "response-media-type-parameter-changed"

	// Parameter changes that cannot be ordered reach this id: parameters
	// that moved in more than one direction at once, or a parameter change
	// arriving together with a change to the media type name itself.
	ResponseMediaTypeNameChangedCommentId = "response-media-type-name-changed-warning-comment"
)

func ResponseMediaTypeNameUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil {
				continue
			}
			if operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for responseStatus, responsesDiff := range operationItem.ResponsesDiff.Modified {
				if responsesDiff.ContentDiff == nil {
					continue
				}
				for _, mediaType := range responsesDiff.ContentDiff.MediaTypeModified {
					if mediaType.NameDiff == nil {
						continue
					}

					fromMediaType, _ := mediaType.NameDiff.NameDiff.From.(string)
					toMediaType, _ := mediaType.NameDiff.NameDiff.To.(string)
					baseSource, revisionSource := responseMediaTypeNameSources(operationsSources, operationItem, responsesDiff, fromMediaType, toMediaType)

					// A difference in the media type parameters, classified by
					// what happened: a parameter appearing narrows what the
					// server may return, one disappearing widens it, and a
					// changed value is neither. Mixed directions cannot be
					// ordered and stay a warning.
					if pd := mediaType.NameDiff.ParametersDiff; !pd.Empty() {
						directions := 0
						for _, n := range []int{len(pd.Added), len(pd.Deleted), len(pd.Modified)} {
							if n > 0 {
								directions++
							}
						}
						nameAlsoChanged := mediaType.NameDiff.TypeDiff != nil ||
							mediaType.NameDiff.SubtypeDiff != nil ||
							mediaType.NameDiff.SuffixDiff != nil
						// the parameter is named separately, so the media type
						// reads as its bare name
						bareName := strings.TrimSpace(strings.SplitN(fromMediaType, ";", 2)[0])
						if directions > 1 || nameAlsoChanged {
							result = append(result, opInfo.NewApiChange(
								ResponseMediaTypeNameChangedId,
								[]any{mediaType.NameDiff.NameDiff.From, mediaType.NameDiff.NameDiff.To, responseStatus},
								ResponseMediaTypeNameChangedCommentId,
							).WithSources(baseSource, revisionSource))
							continue
						}
						for _, param := range pd.Added {
							result = append(result, opInfo.NewApiChange(
								ResponseMediaTypeParameterAddedId,
								[]any{param, bareName, responseStatus},
								"",
							).WithSources(baseSource, revisionSource))
						}
						for _, param := range pd.Deleted {
							result = append(result, opInfo.NewApiChange(
								ResponseMediaTypeParameterRemovedId,
								[]any{param, bareName, responseStatus},
								"",
							).WithSources(baseSource, revisionSource))
						}
						for param, valueDiff := range pd.Modified {
							result = append(result, opInfo.NewApiChange(
								ResponseMediaTypeParameterChangedId,
								[]any{param, bareName, valueDiff.From, valueDiff.To, responseStatus},
								"",
							).WithSources(baseSource, revisionSource))
						}
						continue
					}

					// If params didn't change, check if the media type is a generalization or specialization
					id := ResponseMediaTypeNameGeneralizedId
					if mediaType.NameDiff.IsContained() {
						id = ResponseMediaTypeNameSpecializedId
					}

					result = append(result, opInfo.NewApiChange(
						id,
						[]any{mediaType.NameDiff.NameDiff.From, mediaType.NameDiff.NameDiff.To, responseStatus},
						"",
					).WithSources(baseSource, revisionSource))
				}
			}
		}
	}
	return result
}
