package checker

import (
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
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

// responseSource returns a Source for a specific response status code within an operation.
// Falls back to the "responses" field location if the specific status has no origin data.
func responseSource(operationsSources *diff.OperationsSourcesMap, op *openapi3.Operation, responseStatus string) *Source {
	if op == nil {
		return nil
	}

	if op.Responses != nil {
		if responseRef := op.Responses.Value(responseStatus); responseRef != nil {
			if responseRef.Value != nil && responseRef.Value.Origin != nil {
				return NewSourceFromOrigin(operationsSources, op, responseRef.Value.Origin)
			}
		}
	}

	// Fall back to "responses" field within the operation
	if op.Origin == nil {
		return nil
	}
	return NewSourceFromField(operationsSources, op, op.Origin, "responses")
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
			for _, responseStatus := range operationItem.ResponsesDiff.Deleted {
				status, err := strconv.Atoi(responseStatus)
				if err != nil {
					continue
				}

				if filter(status) {
					baseSource := responseSource(operationsSources, operationItem.Base, responseStatus)
					var revisionSource *Source
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
					var baseSource *Source
					revisionSource := responseSource(operationsSources, operationItem.Revision, responseStatus)
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
