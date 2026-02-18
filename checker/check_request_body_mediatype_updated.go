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

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)

			addedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeAdded
			for _, mediaType := range addedMediaTypes {
				result = append(result, NewApiChange(
					RequestBodyMediaTypeAddedId,
					config,
					[]any{mediaType},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}

			removedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeDeleted
			for _, mediaType := range removedMediaTypes {
				result = append(result, NewApiChange(
					RequestBodyMediaTypeRemovedId,
					config,
					[]any{mediaType},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}
		}
	}
	return result
}
