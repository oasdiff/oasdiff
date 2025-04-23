package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
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
					if mediaType.NameDiff.Empty() {
						continue
					}

					name1 := mediaType.NameDiff.From.(string)
					name2 := mediaType.NameDiff.To.(string)

					id := ResponseMediaTypeNameGeneralizedId
					if diff.IsMediaTypeNameContained(name1, name2) {
						id = ResponseMediaTypeNameSpecializedId
					}

					result = append(result, NewApiChange(
						id,
						config,
						[]any{name1, name2, responseStatus},
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
