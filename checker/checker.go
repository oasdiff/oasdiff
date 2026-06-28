package checker

//go:generate go-localize -input localizations_src -output localizations
import (
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

// CheckBackwardCompatibility runs the checks with level WARN and ERR
func CheckBackwardCompatibility(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap) Changes {
	return CheckBackwardCompatibilityUntilLevel(config, diffReport, operationsSources, WARN)
}

// CheckBackwardCompatibilityUntilLevel runs the checks with level equal or higher than the given level.
// It does not mutate the caller's diffReport; clonePathsDiffForCheck isolates the pipeline's mutation surface.
func CheckBackwardCompatibilityUntilLevel(config *Config, diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, level Level) Changes {
	result := make(Changes, 0)

	if diffReport == nil {
		return result
	}

	diffReport = clonePathsDiffForCheck(diffReport)

	result = applyStabilityLevelPolicy(config, diffReport, result, operationsSources)

	mergeWebhookOperationsIntoPathsDiff(diffReport)

	for _, check := range config.Checks {
		if check == nil {
			continue
		}
		errs := check(diffReport, operationsSources, config)
		result = append(result, errs...)
	}

	// PROTOTYPE (#702 follow-up): declarative post-checker suppression of
	// findings made redundant by a co-located headline finding.
	result = suppressSuperseded(result)

	filteredResult := make(Changes, 0)
	for _, change := range result {
		if config.getLogLevel(change.GetId()) >= level {
			filteredResult = append(filteredResult, change)
		}
	}

	slices.SortFunc(filteredResult, CompareChanges)
	return filteredResult
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
