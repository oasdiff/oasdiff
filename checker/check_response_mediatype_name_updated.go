package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseMediaTypeNameChangedId     = "response-media-type-name-changed"
	ResponseMediaTypeNameGeneralizedId = "response-media-type-name-generalized"
	ResponseMediaTypeNameSpecializedId = "response-media-type-name-specialized"
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
			for responseStatus, responsesDiff := range operationItem.ResponsesDiff.Modified {
				if responsesDiff.ContentDiff == nil {
					continue
				}
				for _, mediaType := range responsesDiff.ContentDiff.MediaTypeModified {
					if mediaType.NameDiff == nil {
						continue
					}

					// If parameters changed, this is a changed media type
					if !mediaType.NameDiff.ParametersDiff.Empty() {
						result = append(result, NewApiChange(
							ResponseMediaTypeNameChangedId,
							config,
							[]any{mediaType.NameDiff.NameDiff.From, mediaType.NameDiff.NameDiff.To, responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						))
						continue
					}

					// If params didn't change, check if the media type is a generalization or specialization
					id := ResponseMediaTypeNameGeneralizedId
					if mediaType.NameDiff.IsContained() {
						id = ResponseMediaTypeNameSpecializedId
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{mediaType.NameDiff.NameDiff.From, mediaType.NameDiff.NameDiff.To, responseStatus},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}
			}
		}
	}
	return result
}
