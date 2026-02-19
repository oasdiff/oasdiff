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

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)

			for _, tag := range operationItem.TagsDiff.Deleted {
				result = append(result, NewApiChange(
					APITagRemovedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					op,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}

			for _, tag := range operationItem.TagsDiff.Added {
				result = append(result, NewApiChange(
					APITagAddedId,
					config,
					[]any{tag},
					"",
					operationsSources,
					op,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}
		}
	}
	return result
}
