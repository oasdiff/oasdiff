package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderAddedId = "response-header-added"
)

// ResponseHeaderAddedCheck reports a response header that was added. Adding a
// response header is non-breaking, so it is reported at INFO. It is the mirror
// of ResponseHeaderRemovedCheck (see #1033).
func ResponseHeaderAddedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.HeadersDiff == nil {
					continue
				}

				for _, headerName := range responseDiff.HeadersDiff.Added {
					if responseDiff.Revision.Headers[headerName] == nil {
						continue
					}
					header := responseDiff.Revision.Headers[headerName].Value
					revisionSource := NewSourceFromOrigin(operationsSources, operationItem.Revision, header.Origin)
					result = append(result, NewApiChange(
						ResponseHeaderAddedId,
						config,
						[]any{headerName, responseStatus},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					).WithSources(nil, revisionSource))
				}
			}
		}
	}
	return result
}
