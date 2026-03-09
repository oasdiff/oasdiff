package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequiredResponseHeaderRemovedId = "required-response-header-removed"
	OptionalResponseHeaderRemovedId = "optional-response-header-removed"
)

func ResponseHeaderRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

				for _, headerName := range responseDiff.HeadersDiff.Deleted {
					if responseDiff.Base.Headers[headerName] == nil {
						continue
					}
					header := responseDiff.Base.Headers[headerName].Value
					baseSource := NewSourceFromOrigin(operationsSources, operationItem.Base, header.Origin)
					required := header.Required
					if required {
						result = append(result, NewApiChange(
							RequiredResponseHeaderRemovedId,
							config,
							[]any{headerName, responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil))
					} else {
						result = append(result, NewApiChange(
							OptionalResponseHeaderRemovedId,
							config,
							[]any{headerName, responseStatus},
							"",
							operationsSources,
							operationItem.Revision,
							operation,
							path,
						).WithSources(baseSource, nil))
					}
				}
			}
		}
	}
	return result
}
