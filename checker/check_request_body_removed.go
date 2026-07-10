package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyRemovedId = "request-body-removed"
)

func RequestBodyRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}

		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil {
				continue
			}

			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)

			if operationItem.RequestBodyDiff.Deleted {
				baseSource := requestBodySource(operationsSources, operationItem.Base)
				result = append(result, opInfo.NewApiChange(
					RequestBodyRemovedId,
					nil,
					"",
				).WithSources(baseSource, nil))
			}
		}
	}
	return result
}
