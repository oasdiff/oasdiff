package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APITagRemovedId = "api-tag-removed"
	APITagAddedId   = "api-tag-added"
)

func APITagUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}

		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			op := pathItem.Base.GetOperation(operation)

			if operationItem.TagsDiff == nil {
				continue
			}

			for _, tag := range operationItem.TagsDiff.Deleted {
				var baseSource *Source
				if operationItem.Base != nil && operationItem.Base.Origin != nil {
					baseSource = NewSourceFromSequenceItem(operationsSources, operationItem.Base, operationItem.Base.Origin, "tags", tag)
				}
				result = append(result, NewApiChange(
					APITagRemovedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					op,
					operation,
					path,
				).WithSources(baseSource, nil))
			}

			for _, tag := range operationItem.TagsDiff.Added {
				var revisionSource *Source
				if operationItem.Revision != nil && operationItem.Revision.Origin != nil {
					revisionSource = NewSourceFromSequenceItem(operationsSources, operationItem.Revision, operationItem.Revision.Origin, "tags", tag)
				}
				result = append(result, NewApiChange(
					APITagAddedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					op,
					operation,
					path,
				).WithSources(nil, revisionSource))
			}
		}
	}
	return result
}
