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
)

// CheckBackwardCompatibility runs the checks with level WARN and ERR
func CheckBackwardCompatibility(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	return CheckBackwardCompatibilityUntilLevel(config, diffReport, operationsSources, WARN)
}

// CheckBackwardCompatibilityUntilLevel runs the checks with level equal or higher than the given level.
// The caller's diffReport is not mutated — internally we clone the
// PathsDiff and the nested maps/slices that the check pipeline writes
// to (the rest of the diff is shared by reference, since checks only
// read it).
func CheckBackwardCompatibilityUntilLevel(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, level Level) Changes {
	result := make(Changes, 0)

	if diffReport == nil {
		return result
	}

	diffReport = clonePathsDiffForCheck(diffReport)

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
			if baseStability != STABILITY_DRAFT && baseStability != STABILITY_ALPHA {
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
			if baseStability != STABILITY_DRAFT && baseStability != STABILITY_ALPHA {
				pathDiff.OperationsDiff.Deleted[iOperation] = operation
				iOperation++
			}
		}
		pathDiff.OperationsDiff.Deleted = pathDiff.OperationsDiff.Deleted[:iOperation]

		// remove draft and alpha operations diffs modified
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
			if baseStability == STABILITY_STABLE && revisionStability != STABILITY_STABLE ||
				baseStability == STABILITY_BETA && revisionStability != STABILITY_BETA && revisionStability != STABILITY_STABLE ||
				baseStability == STABILITY_ALPHA && revisionStability != STABILITY_ALPHA && revisionStability != STABILITY_BETA && revisionStability != STABILITY_STABLE ||
				revisionStability == "" && baseStability != "" {
				result = append(result, NewApiChange(
					APIStabilityDecreasedId,
					config,
					[]any{baseStability, revisionStability},
					"",
					operationsSources,
					operationItem.Revision,
					operation,
					path,
				))

				continue
			}
			if revisionStability == STABILITY_DRAFT || revisionStability == STABILITY_ALPHA {
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

// clonePathsDiffForCheck returns a copy of diffReport whose PathsDiff —
// and the specific nested fields the check pipeline writes to — are
// separate allocations from the caller's. Everything outside that
// mutation surface (WebhooksDiff, ComponentsDiff, etc., and the
// MethodDiff/PathDiff values themselves, which are only read) is
// shared by reference.
//
// Mutation surface (must be cloned):
//   - PathsDiff struct itself (Deleted/Modified are reassigned)
//   - PathsDiff.Deleted slice (truncated in-place)
//   - PathsDiff.Modified map (webhook entries inserted, see
//     mergeWebhookOperationsIntoPathsDiff)
//   - For each PathDiff in PathsDiff.Modified: a fresh PathDiff
//     because OperationsDiff.Deleted gets truncated and
//     OperationsDiff.Modified has keys deleted.
func clonePathsDiffForCheck(d *diff.Diff) *diff.Diff {
	cloned := *d
	if d.PathsDiff == nil {
		return &cloned
	}

	pdCopy := *d.PathsDiff
	cloned.PathsDiff = &pdCopy
	cloned.PathsDiff.Deleted = slices.Clone(d.PathsDiff.Deleted)

	if d.PathsDiff.Modified != nil {
		cloned.PathsDiff.Modified = make(diff.ModifiedPaths, len(d.PathsDiff.Modified))
		for k, v := range d.PathsDiff.Modified {
			pdElem := *v
			if v.OperationsDiff != nil {
				odCopy := *v.OperationsDiff
				pdElem.OperationsDiff = &odCopy
				pdElem.OperationsDiff.Deleted = slices.Clone(v.OperationsDiff.Deleted)
				if v.OperationsDiff.Modified != nil {
					pdElem.OperationsDiff.Modified = make(diff.ModifiedOperations, len(v.OperationsDiff.Modified))
					for ok, ov := range v.OperationsDiff.Modified {
						pdElem.OperationsDiff.Modified[ok] = ov
					}
				}
			}
			cloned.PathsDiff.Modified[k] = &pdElem
		}
	}

	return &cloned
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
