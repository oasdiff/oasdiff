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

	mergeWebhookOperationsIntoPathsDiff(diffReport)

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
	return applyStabilityLevelPolicy(config, diffReport, result, operationsSources)
}

// applyStabilityLevelPolicy is a thin pipeline that validates, detects changes, and filters.
func applyStabilityLevelPolicy(config *Config, diffReport *diff.Diff, result Changes, operationsSources *diff.OperationsSourcesMap) Changes {
	if diffReport.PathsDiff == nil {
		return result
	}
	result = append(result, checkInvalidStabilityLevels(config, diffReport, operationsSources)...)
	result = append(result, checkStabilityLevelChanged(config, diffReport, operationsSources)...)
	filterBelowStabilityThreshold(config, diffReport)
	return result
}

// normalizedStability returns STABILITY_STABLE for an empty label, otherwise the label itself.
func normalizedStability(label string) string {
	if label == "" {
		return STABILITY_STABLE
	}
	return label
}

// lowerStability returns the stability label with the lower level (by ParseStabilityLevel ordering).
func lowerStability(baseLabel, revisionLabel string) string {
	if ParseStabilityLevel(baseLabel) <= ParseStabilityLevel(revisionLabel) {
		return baseLabel
	}
	return revisionLabel
}

// checkInvalidStabilityLevels emits change records for unparseable x-stability-level values.
// It does not mutate diffReport.
func checkInvalidStabilityLevels(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	var result Changes

	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathDiff.OperationsDiff.Modified {
			if _, err := getStabilityLevel(pathDiff.Base.GetOperation(operation).Extensions); err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Base, operationsSources, operation, path, err))
			}
			if _, err := getStabilityLevel(pathDiff.Revision.GetOperation(operation).Extensions); err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Revision, operationsSources, operation, path, err))
			}
		}
	}

	// Also check deleted paths/operations for invalid stability
	for _, path := range diffReport.PathsDiff.Deleted {
		pathVal := diffReport.PathsDiff.Base.Value(path)
		for operation, operationItem := range pathVal.Operations() {
			if _, err := getStabilityLevel(pathVal.GetOperation(operation).Extensions); err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err))
			}
		}
	}

	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		for _, operation := range pathDiff.OperationsDiff.Deleted {
			operationItem := pathDiff.Base.GetOperation(operation)
			if _, err := getStabilityLevel(operationItem.Extensions); err != nil {
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err))
			}
		}
	}

	return result
}

// checkStabilityLevelChanged emits APIStabilityDecreasedId / APIStabilityIncreasedId changes.
// It does not mutate diffReport.
func checkStabilityLevelChanged(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	var result Changes

	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathDiff.OperationsDiff.Modified {
			baseStability, err := getStabilityLevel(pathDiff.Base.GetOperation(operation).Extensions)
			if err != nil {
				continue // already reported by checkInvalidStabilityLevels
			}
			revisionStability, err := getStabilityLevel(pathDiff.Revision.GetOperation(operation).Extensions)
			if err != nil {
				continue // already reported by checkInvalidStabilityLevels
			}

			baseLabel := normalizedStability(baseStability)
			revisionLabel := normalizedStability(revisionStability)
			baseLevel := ParseStabilityLevel(baseLabel)
			revisionLevel := ParseStabilityLevel(revisionLabel)

			if baseLevel == revisionLevel {
				continue
			}

			if config != nil && !config.StabilityLevel.IsIncluded(lowerStability(baseLabel, revisionLabel)) {
				continue
			}

			changeId := APIStabilityIncreasedId
			if revisionLevel < baseLevel {
				changeId = APIStabilityDecreasedId
			}

			result = append(result, NewApiChange(
				changeId,
				config,
				[]any{baseLabel, revisionLabel},
				"",
				operationsSources,
				operationItem.Revision,
				operation,
				path,
			))
		}
	}

	return result
}

// filterBelowStabilityThreshold mutates diffReport by removing paths and operations
// that fall below the configured stability threshold. It emits no Changes.
func filterBelowStabilityThreshold(config *Config, diffReport *diff.Diff) {
	if config == nil {
		return
	}

	// Filter deleted paths
	iPath := 0
	for _, path := range diffReport.PathsDiff.Deleted {
		keep := false
		pathVal := diffReport.PathsDiff.Base.Value(path)
		for operation := range pathVal.Operations() {
			baseStability, err := getStabilityLevel(pathVal.GetOperation(operation).Extensions)
			if err != nil {
				continue
			}
			if config.StabilityLevel.IsIncluded(baseStability) {
				keep = true
				break
			}
		}
		if keep {
			diffReport.PathsDiff.Deleted[iPath] = path
			iPath++
		}
	}
	diffReport.PathsDiff.Deleted = diffReport.PathsDiff.Deleted[:iPath]

	// Filter modified paths → deleted operations and modified operations
	for _, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}

		// Filter deleted operations
		iOperation := 0
		for _, operation := range pathDiff.OperationsDiff.Deleted {
			operationItem := pathDiff.Base.GetOperation(operation)
			baseStability, err := getStabilityLevel(operationItem.Extensions)
			if err != nil {
				continue
			}
			if config.StabilityLevel.IsIncluded(baseStability) {
				pathDiff.OperationsDiff.Deleted[iOperation] = operation
				iOperation++
			}
		}
		pathDiff.OperationsDiff.Deleted = pathDiff.OperationsDiff.Deleted[:iOperation]

		// Filter modified operations below threshold
		for operation := range pathDiff.OperationsDiff.Modified {
			revisionStability, err := getStabilityLevel(pathDiff.Revision.GetOperation(operation).Extensions)
			if err != nil {
				continue
			}
			if !config.StabilityLevel.IsIncluded(revisionStability) {
				delete(pathDiff.OperationsDiff.Modified, operation)
			}
		}
	}
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

// mergeWebhookOperationsIntoPathsDiff merges modified webhooks into PathsDiff.Modified with a "webhook:" prefix.
// This allows all existing path/operation checker rules to apply to webhooks without code duplication.
// Only Modified webhooks are merged because each entry in Modified is self-contained with its own Base/Revision data.
// Added/Deleted webhooks cannot be merged because PathsDiff.Added/Deleted are just path name strings that require
// lookup in PathsDiff.Base/Revision (openapi3.Paths objects that don't contain webhooks).
// Added/Deleted webhooks are handled separately by WebhookUpdatedCheck.
func mergeWebhookOperationsIntoPathsDiff(diffReport *diff.Diff) {
	if diffReport.WebhooksDiff == nil || len(diffReport.WebhooksDiff.Modified) == 0 {
		return
	}
	if diffReport.PathsDiff == nil {
		diffReport.PathsDiff = &diff.PathsDiff{Modified: diff.ModifiedPaths{}}
	}
	if diffReport.PathsDiff.Modified == nil {
		diffReport.PathsDiff.Modified = diff.ModifiedPaths{}
	}
	for name, pathDiff := range diffReport.WebhooksDiff.Modified {
		diffReport.PathsDiff.Modified["webhook:"+name] = pathDiff
	}
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
