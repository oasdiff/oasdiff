package checker

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APIStabilityDecreasedId = "api-stability-decreased"
	APIStabilityIncreasedId = "api-stability-increased"
)

// applyStabilityLevelPolicy is a thin pipeline that validates, detects changes, and filters.
func applyStabilityLevelPolicy(config *Config, diffReport *diff.Diff, result Changes, operationsSources *diff.OperationsSourcesMap) Changes {
	if diffReport.PathsDiff == nil {
		return result
	}
	result = append(result, checkInvalidStabilityLevels(config, diffReport, operationsSources)...)
	result = append(result, checkStabilityLevelChanged(config, diffReport, operationsSources)...)
	filterBelowStabilityThreshold(config, diffReport)
	// Detect property-level stability changes after filtering, so a property
	// change on an operation below the threshold (already removed from the diff)
	// is not reported. Unlike an endpoint stability change, a property change
	// never moves its operation out of scope, so an in-scope operation survives
	// the filter and its property changes are still seen here.
	result = append(result, RequestPropertyStabilityUpdatedCheck(diffReport, operationsSources, config)...)
	result = append(result, ResponsePropertyStabilityUpdatedCheck(diffReport, operationsSources, config)...)
	return result
}

// checkInvalidStabilityLevels emits change records for unparseable x-stability-level values.
func checkInvalidStabilityLevels(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	var result Changes

	for path, pathDiff := range diffReport.PathsDiff.Modified {
		if pathDiff.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathDiff.OperationsDiff.Modified {
			if _, err := getStabilityLevel(pathDiff.Base.GetOperation(operation).Extensions); err != nil {
				baseSource := stabilityFieldSource(operationsSources, operationItem.Base, operationItem.Base.Origin)
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Base, operationsSources, operation, path, err).WithSources(baseSource, nil))
			}
			if _, err := getStabilityLevel(pathDiff.Revision.GetOperation(operation).Extensions); err != nil {
				revisionSource := stabilityFieldSource(operationsSources, operationItem.Revision, operationItem.Revision.Origin)
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem.Revision, operationsSources, operation, path, err).WithSources(nil, revisionSource))
			}
		}
		for _, operation := range pathDiff.OperationsDiff.Deleted {
			operationItem := pathDiff.Base.GetOperation(operation)
			if _, err := getStabilityLevel(operationItem.Extensions); err != nil {
				baseSource := stabilityFieldSource(operationsSources, operationItem, operationItem.Origin)
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err).WithSources(baseSource, nil))
			}
		}
	}

	// Also check deleted paths for invalid stability
	for _, path := range diffReport.PathsDiff.Deleted {
		pathVal := diffReport.PathsDiff.Base.Value(path)
		for operation, operationItem := range pathVal.Operations() {
			if _, err := getStabilityLevel(pathVal.GetOperation(operation).Extensions); err != nil {
				baseSource := stabilityFieldSource(operationsSources, operationItem, operationItem.Origin)
				result = append(result, getAPIInvalidStabilityLevel(config, operationItem, operationsSources, operation, path, err).WithSources(baseSource, nil))
			}
		}
	}

	return result
}

// checkStabilityLevelChanged emits APIStabilityDecreasedId / APIStabilityIncreasedId changes.
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

			// Gate on the base stability (the level the element is leaving), not the
			// lower of the two. This reports changes to elements that were within the
			// configured threshold before the change: a stable->draft destabilization
			// is reported (base stable meets the threshold), while a draft->stable
			// change of a previously out-of-scope element is not. Using the lower
			// level instead would drop stable->draft by default, which regresses the
			// existing api-stability-decreased ERR.
			if config != nil && !config.StabilityLevel.IsIncluded(baseLabel) {
				continue
			}

			changeId := APIStabilityIncreasedId
			if revisionLevel < baseLevel {
				changeId = APIStabilityDecreasedId
			}

			baseSource := stabilityFieldSource(operationsSources, pathDiff.Base.GetOperation(operation), pathDiff.Base.GetOperation(operation).Origin)
			revisionSource := stabilityFieldSource(operationsSources, operationItem.Revision, operationItem.Revision.Origin)

			result = append(result, NewApiChange(
				changeId,
				config,
				[]any{baseLabel, revisionLabel},
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

// normalizedStability returns StabilityStable for an empty label, otherwise the label itself.
func normalizedStability(label string) string {
	if label == "" {
		return StabilityStable
	}
	return label
}

func getAPIInvalidStabilityLevel(config *Config, operation *openapi3.Operation, operationsSources *diff.OperationsSourcesMap, method string, path string, err error) ApiChange {
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

	if stabilityLevel != StabilityDraft &&
		stabilityLevel != StabilityAlpha &&
		stabilityLevel != StabilityBeta &&
		stabilityLevel != StabilityStable {
		return "", fmt.Errorf("value is not one of %s, %s, %s or %s: %q", StabilityDraft, StabilityAlpha, StabilityBeta, StabilityStable, stabilityLevel)
	}

	return stabilityLevel, nil
}
