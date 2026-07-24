package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderBecameNullableId    = "response-header-became-nullable"
	ResponseHeaderBecameNotNullableId = "response-header-became-not-nullable"
)

// ResponseHeaderBecameNullableCheck reports a nullability change on a response
// header's schema, the response-header mirror of ResponsePropertyBecameNullableCheck.
// A response header becoming nullable means a client that never expected a null
// value can now receive one (breaking); becoming not-nullable narrows the
// server's output and is safe (info).
func ResponseHeaderBecameNullableCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}
			opInfo := newOpInfoFromDiff(config, operationItem, operationsSources, operation, path)
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.HeadersDiff == nil {
					continue
				}
				for headerName, headerDiff := range responseDiff.HeadersDiff.Modified {
					if id := nullabilityChangeId(headerDiff.SchemaDiff, ResponseHeaderBecameNullableId, ResponseHeaderBecameNotNullableId); id != "" {
						baseSource, revisionSource := headerSources(operationsSources, operationItem, responseDiff, headerName)
						result = append(result, opInfo.NewApiChange(
							id,
							[]any{headerName, responseStatus},
							"",
						).WithSources(baseSource, revisionSource))
					}
				}
			}
		}
	}
	return result
}
