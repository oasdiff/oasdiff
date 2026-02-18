package checker

import (
	"strconv"
	"strings"

	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseSuccessStatusRemovedId    = "response-success-status-removed"
	ResponseNonSuccessStatusRemovedId = "response-non-success-status-removed"
	ResponseSuccessStatusAddedId      = "response-success-status-added"
	ResponseNonSuccessStatusAddedId   = "response-non-success-status-added"
)

func ResponseSuccessStatusUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	success := func(status int) bool {
		return status >= 200 && status <= 299
	}

	return responseStatusUpdated(diffReport, operationsSources, config, success, ResponseSuccessStatusRemovedId)
}

func ResponseNonSuccessStatusUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	notSuccess := func(status int) bool {
		return status < 200 || status > 299
	}

	return responseStatusUpdated(diffReport, operationsSources, config, notSuccess, ResponseNonSuccessStatusRemovedId)
}

func responseStatusUpdated(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config, filter func(int) bool, id string) Changes {
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
			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			for _, responseStatus := range operationItem.ResponsesDiff.Deleted {
				status, err := strconv.Atoi(responseStatus)
				if err != nil {
					continue
				}

				if filter(status) {
					result = append(result, NewApiChange(
						id,
						config,
						[]any{responseStatus},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithSources(baseSource, revisionSource))
				}
			}

			for _, responseStatus := range operationItem.ResponsesDiff.Added {
				addedId := strings.Replace(id, "removed", "added", 1)
				status, err := strconv.Atoi(responseStatus)
				if err != nil {
					continue
				}

				if filter(status) {
					result = append(result, NewApiChange(
						addedId,
						config,
						[]any{responseStatus},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithSources(baseSource, revisionSource))
				}
			}
		}
	}
	return result
}
