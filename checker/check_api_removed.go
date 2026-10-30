package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
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
			if err != nil || stability == StabilityAlpha || stability == StabilityDraft {
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
	id, extra, ok := removalId(opInfo.operation.Deprecated, opInfo.operation.Extensions, removalIds{
		removed:         getWithoutDeprecationId(isPath),
		withDeprecation: getWithDeprecationId(isPath),
		beforeSunset:    getBeforeSunsetId(isPath),
		sunsetParse:     APIPathSunsetParseId,
	})
	if !ok {
		return nil
	}

	baseSource := location.NewSourceFromOrigin(opInfo.operationsSources, opInfo.operation, opInfo.operation.Origin)
	return opInfo.NewApiChange(id, extra, "").WithSources(baseSource, location.NewEmptySource())
}

func getAPIPathSunsetParse(opInfo opInfo, err error) ApiChange {
	return opInfo.NewApiChange(
		APIPathSunsetParseId,
		[]any{err},
		"",
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
