package checker

//go:generate go-localize -input localizations_src -output localizations
import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APIStabilityDecreasedId = "api-stability-decreased"
	APIStabilityIncreasedId = "api-stability-increased"
)

// CheckBackwardCompatibility runs the checks with level WARN and ERR
func CheckBackwardCompatibility(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	return CheckBackwardCompatibilityUntilLevel(config, diffReport, operationsSources, WARN)
}

// CheckBackwardCompatibilityUntilLevel runs the checks with level equal or higher than the given level
func CheckBackwardCompatibilityUntilLevel(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, level Level) Changes {
	result := make(Changes, 0)

	if diffReport == nil {
		return result
	}

	result = removeDraftAndAlphaOperationsDiffs(config, diffReport, result, operationsSources)

	for _, check := range config.Checks {
		if check == nil {
			continue
		}
		errs := check(diffReport, operationsSources, config)
		result = append(result, errs...)
	}

	filteredResult := make(Changes, 0)
	for _, change := range result {
		if config.getLogLevel(change.GetId()) >= level {
			filteredResult = append(filteredResult, change)
		}
	}

	slices.SortFunc(filteredResult, CompareChanges)
	return filteredResult
}

func removeDraftAndAlphaOperationsDiffs(config *Config, diffReport *diff.Diff, result Changes, operationsSources *diff.OperationsSourcesMap) Changes {
	if diffReport.PathsDiff == nil {
		return result
	}
	// remove draft and alpha paths diffs delete
	iPath := 0
	for _, path := range diffReport.PathsDiff.Deleted {
		ignore := true
		pathDiff := diffReport.PathsDiff
		for operation, operationItem := range pathDiff.Base.Value(path).Operations() {
			baseStability, err := getStabilityLevel(pathDiff.Base.Value(path).GetOperation(operation).Extensions)
			if err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err))
				continue
			}
			if config != nil && config.StabilityLevel.IsIncluded(baseStability) {
				ignore = false
				break
			}
		}
		if !ignore {
			diffReport.PathsDiff.Deleted[iPath] = path
			iPath++
		}
	}
	diffReport.PathsDiff.Deleted = diffReport.PathsDiff.Deleted[:iPath]

	// remove draft and alpha paths diffs modified
	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		// remove draft and alpha operations diffs deleted
		iOperation := 0
		for _, operation := range pathDiff.OperationsDiff.Deleted {
			operationItem := pathDiff.Base.GetOperation(operation)
			baseStability, err := getStabilityLevel(operationItem.Extensions)
			if err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err))
				continue
			}
			if config != nil && config.StabilityLevel.IsIncluded(baseStability) {
				pathDiff.OperationsDiff.Deleted[iOperation] = operation
				iOperation++
			}
		}
		pathDiff.OperationsDiff.Deleted = pathDiff.OperationsDiff.Deleted[:iOperation]

		// Always check for invalid stability levels in modified operations
		for operation, operationItem := range pathDiff.OperationsDiff.Modified {
			baseStability, err := getStabilityLevel(pathDiff.Base.GetOperation(operation).Extensions)
			if err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Base, operationsSources, operation, path, err))
				continue
			}
			revisionStability, err := getStabilityLevel(pathDiff.Revision.GetOperation(operation).Extensions)
			if err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Revision, operationsSources, operation, path, err))
				continue
			}

			// Normalize: no x-stability-level is treated as implicit "stable"
			baseLabel := baseStability
			if baseLabel == "" {
				baseLabel = STABILITY_STABLE
			}
			revisionLabel := revisionStability
			if revisionLabel == "" {
				revisionLabel = STABILITY_STABLE
			}

			baseLevel := ParseStabilityLevel(baseLabel)
			revisionLevel := ParseStabilityLevel(revisionLabel)

			// Default behavior (no flag): only detect stable→beta decrease
			if baseLabel == STABILITY_STABLE && revisionLabel == STABILITY_BETA {
				result = append(result, NewApiChange(
					APIStabilityDecreasedId,
					config,
					[]any{baseLabel, revisionLabel},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				))
			}

			if config == nil || config.StabilityLevel == StabilityLevelNone {
				// Default behavior: filter out draft and alpha operations
				if revisionStability == STABILITY_DRAFT || revisionStability == STABILITY_ALPHA {
					delete(pathDiff.OperationsDiff.Modified, operation)
				}
				continue
			}

			// With --stability-level flag: detect all stability changes at/above threshold
			// (skip stable→beta since already reported above)
			if revisionLevel < baseLevel && !(baseLabel == STABILITY_STABLE && revisionLabel == STABILITY_BETA) {
				if config.StabilityLevel.IsIncluded(baseLabel) || config.StabilityLevel.IsIncluded(revisionLabel) {
					result = append(result, NewApiChange(
						APIStabilityDecreasedId,
						config,
						[]any{baseLabel, revisionLabel},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}
				continue
			}
			if revisionLevel > baseLevel {
				// Stability increased (e.g., alpha→beta, draft→stable)
				if config.StabilityLevel.IsIncluded(baseLabel) || config.StabilityLevel.IsIncluded(revisionLabel) {
					result = append(result, NewApiChange(
						APIStabilityIncreasedId,
						config,
						[]any{baseLabel, revisionLabel},
						"",
						operationsSources,
						operationItem.Revision,
						operation,
						path,
					))
				}
				continue
			}
			// Filter out operations below the configured stability threshold
			// Use the higher of base/revision stability to decide inclusion
			effectiveStability := baseStability
			if revisionStability != "" && ParseStabilityLevel(revisionStability) < ParseStabilityLevel(effectiveStability) {
				effectiveStability = revisionStability
			}
			if !config.StabilityLevel.IsIncluded(effectiveStability) {
				delete(pathDiff.OperationsDiff.Modified, operation)
			}
		}
	}
	return result
}

func getAPIInvalidStabilityLevel(config *Config, operation *openapi3.Operation, operationsSources *diff.OperationsSourcesMap, method string, path string, err error) Change {
	return NewApiChange(
		APIInvalidStabilityLevelId,
		config,
		[]any{err},
		"",
		operationsSources,
		operation,
		method,
		path,
	)
}

func getStabilityLevel(i map[string]any) (string, error) {
	if i == nil || i[diff.XStabilityLevelExtension] == nil {
		return "", nil
	}
	var stabilityLevel string

	stabilityLevel, ok := i[diff.XStabilityLevelExtension].(string)
	if !ok {
		jsonStabilityRaw, ok := i[diff.XStabilityLevelExtension].(json.RawMessage)
		if !ok {
			return "", fmt.Errorf("x-stability-level isn't a string nor valid json")
		}
		err := json.Unmarshal(jsonStabilityRaw, &stabilityLevel)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal x-stability-level json")
		}
	}

	if stabilityLevel != STABILITY_DRAFT &&
		stabilityLevel != STABILITY_ALPHA &&
		stabilityLevel != STABILITY_BETA &&
		stabilityLevel != STABILITY_STABLE {
		return "", fmt.Errorf("value is not one of %s, %s, %s or %s: %q", STABILITY_DRAFT, STABILITY_ALPHA, STABILITY_BETA, STABILITY_STABLE, stabilityLevel)
	}

	return stabilityLevel, nil
}
