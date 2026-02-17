package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APIPathRemovedWithoutDeprecationId = "api-path-removed-without-deprecation"
	APIPathRemovedWithDeprecationId    = "api-path-removed-with-deprecation"
	APIPathSunsetParseId               = "api-path-sunset-parse"
	APIPathRemovedBeforeSunsetId       = "api-path-removed-before-sunset"
	APIRemovedWithoutDeprecationId     = "api-removed-without-deprecation"
	APIRemovedWithDeprecationId        = "api-removed-with-deprecation"
	APIRemovedBeforeSunsetId           = "api-removed-before-sunset"
)

func APIRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	return append(
		checkRemovedPaths(diffReport.PathsDiff, operationsSources, config),
		checkRemovedOperations(diffReport.PathsDiff, operationsSources, config)...,
	)
}

func checkRemovedPaths(pathsDiff *diff.PathsDiff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {

	if pathsDiff == nil {
		return nil
	}

	result := make(Changes, 0)
	for _, path := range pathsDiff.Deleted {
		if pathsDiff.Base.Value(path) == nil {
			continue
		}

		for operation := range pathsDiff.Base.Value(path).Operations() {
			op := pathsDiff.Base.Value(path).GetOperation(operation)
			stability, err := getStabilityLevel(op.Extensions)
			if err != nil || stability == STABILITY_ALPHA || stability == STABILITY_DRAFT {
				continue
			}

			opInfo := newOpInfo(config, op, operationsSources, operation, path)
			if change := checkAPIRemoval(opInfo, true); change != nil {
				result = append(result, change)
			}
		}
	}
	return result
}

func checkRemovedOperations(pathsDiff *diff.PathsDiff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	if pathsDiff == nil {
		return nil
	}

	result := make(Changes, 0)

	for path, pathItem := range pathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for _, operation := range pathItem.OperationsDiff.Deleted {
			opInfo := newOpInfo(config, pathItem.Base.GetOperation(operation), operationsSources, operation, path)
			if change := checkAPIRemoval(opInfo, false); change != nil {
				result = append(result, change)
			}
		}
	}

	return result
}

func checkAPIRemoval(opInfo opInfo, isPath bool) Change {

	baseSource := NewSourceFromOrigin(opInfo.operationsSources, opInfo.operation, opInfo.operation.Origin)
	revisionSource := NewEmptySource()

	if !opInfo.operation.Deprecated {
		return NewApiChange(
			getWithoutDeprecationId(isPath),
			opInfo.config,
			nil,
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, revisionSource)
	}
	sunset, ok := getSunset(opInfo.operation.Extensions)
	if !ok {
		return NewApiChange(
			getWithDeprecationId(isPath),
			opInfo.config,
			nil,
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, revisionSource)
	}

	date, err := getSunsetDate(sunset)
	if err != nil {
		return NewApiChange(
			APIPathSunsetParseId,
			opInfo.config,
			[]any{err},
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, nil)
	}

	if civil.DateOf(time.Now()).Before(date) {
		return NewApiChange(
			getBeforeSunsetId(isPath),
			opInfo.config,
			[]any{date},
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		).WithSources(baseSource, revisionSource)
	}
	return nil
}

func getAPIPathSunsetParse(opInfo opInfo, err error) Change {
	return NewApiChange(
		APIPathSunsetParseId,
		opInfo.config,
		[]any{err},
		"",
		opInfo.operationsSources,
		opInfo.operation,
		opInfo.method,
		opInfo.path,
	)
}

func getWithDeprecationId(isPath bool) string {
	if isPath {
		return APIPathRemovedWithDeprecationId
	}
	return APIRemovedWithDeprecationId
}

func getWithoutDeprecationId(isPath bool) string {
	if isPath {
		return APIPathRemovedWithoutDeprecationId
	}
	return APIRemovedWithoutDeprecationId
}

func getBeforeSunsetId(isPath bool) string {
	if isPath {
		return APIPathRemovedBeforeSunsetId
	}
	return APIRemovedBeforeSunsetId
}
