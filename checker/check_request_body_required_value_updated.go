package checker

import (
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestBodyBecameOptionalId = "request-body-became-optional"
	RequestBodyBecameRequiredId = "request-body-became-required"
)

func RequestBodyRequiredUpdatedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
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

			if operationItem.RequestBodyDiff.RequiredDiff == nil {
				continue
			}

			id := RequestBodyBecameOptionalId
			if operationItem.RequestBodyDiff.RequiredDiff.To == true {
				id = RequestBodyBecameRequiredId
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)
			result = append(result, NewApiChange(
				id,
				config,
				nil,
				"",
				operationsSources,
				operationItem.Revision,
				operation,
				path,
			).WithSources(baseSource, revisionSource))
		}
	}
	return result
}
