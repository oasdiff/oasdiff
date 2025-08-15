package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	EndpointAddedId = "endpoint-added"
)

func APIAddedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	appendErr := func(path, method string, operation *openapi3.Operation) {
		result = append(result, NewApiChangeWithSources(
			EndpointAddedId,
			config,
			nil,
			"",
			operationsSources,
			operation,
			method,
			path,
			NewEmptySource(),
			NewSourceFromOrigin(operationsSources, operation, operation.Origin),
		))
	}

	for _, path := range diffReport.PathsDiff.Added {
		for opName, op := range diffReport.PathsDiff.Revision.Value(path).Operations() {
			appendErr(path, opName, op)
		}
	}

	for path, pathDiff := range diffReport.PathsDiff.Modified {
		for opName, op := range pathDiff.Revision.Operations() {
			if baseOp := pathDiff.Base.GetOperation(opName); baseOp != nil {
				continue
			}
			appendErr(path, opName, op)
		}
	}

	return result
}
