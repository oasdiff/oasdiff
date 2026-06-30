package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyMediaTypeAddedId   = "request-body-media-type-added"
	RequestBodyMediaTypeRemovedId = "request-body-media-type-removed"
)

func RequestBodyMediaTypeChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)

			addedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeAdded
			for _, mediaType := range addedMediaTypes {
				revisionSource := requestBodyMediaTypeSource(operationsSources, operationItem.Revision, mediaType)
				result = append(result, opInfo.NewApiChange(
					RequestBodyMediaTypeAddedId,
					[]any{mediaType},
					"",
				).WithSources(nil, revisionSource))
			}

			removedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeDeleted
			for _, mediaType := range removedMediaTypes {
				baseSource := requestBodyMediaTypeSource(operationsSources, operationItem.Base, mediaType)
				result = append(result, opInfo.NewApiChange(
					RequestBodyMediaTypeRemovedId,
					[]any{mediaType},
					"",
				).WithSources(baseSource, nil))
			}
		}
	}
	return result
}
