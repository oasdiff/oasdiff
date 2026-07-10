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
			if operationItem.TagsDiff == nil {
				continue
			}

			baseOp := pathItem.Base.GetOperation(operation)
			revisionOp := pathItem.Revision.GetOperation(operation)
			fieldBase, fieldRevision := operationFieldSources(operationsSources, operationItem, "tags")

			for _, tag := range operationItem.TagsDiff.Deleted {
				result = append(result, NewApiChange(
					APITagRemovedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					baseOp,
					operation,
					path,
				).WithSources(sequenceItemSource(operationsSources, baseOp, "tags", tag, fieldBase), nil))
			}

			for _, tag := range operationItem.TagsDiff.Added {
				result = append(result, NewApiChange(
					APITagAddedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					baseOp,
					operation,
					path,
				).WithSources(nil, sequenceItemSource(operationsSources, revisionOp, "tags", tag, fieldRevision)))
			}
		}
	}
	return result
}
