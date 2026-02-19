package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponseHeaderBecameOptionalId = "response-header-became-optional"
)

func ResponseHeaderBecameOptionalCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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
			for responseStatus, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff.HeadersDiff == nil {
					continue
				}

				for headerName, headerDiff := range responseDiff.HeadersDiff.Modified {
					requiredDiff := headerDiff.RequiredDiff
					if requiredDiff == nil {
						continue
					}
					if requiredDiff.From != true {
						continue
					}

					result = append(result, NewApiChange(
						ResponseHeaderBecameOptionalId,
						config,
						[]any{headerName, responseStatus},
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
