package checker

//go:generate go-localize -input localizations_src -output localizations
import (
	"maps"
	"slices"

	"github.com/oasdiff/oasdiff/diff"
)

// CheckBackwardCompatibility runs the checks and returns the changes with
// level WARN or ERR; see CheckBackwardCompatibilityUntilLevel.
func CheckBackwardCompatibility(config *Config, diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap) Changes {
	return CheckBackwardCompatibilityUntilLevel(config, diffReport, operationsSources, WARN)
}

// CheckBackwardCompatibilityUntilLevel runs the checks and returns the changes
// with level equal or higher than the given level. Changes claimed by a
// recognized schema transition are dropped, so each transition is reported
// only by its own finding (see transition_claims.go).
// It does not modify the caller's diffReport: the pipeline works on a shallow
// copy and replaces the parts it filters or extends with copies.
func CheckBackwardCompatibilityUntilLevel(config *Config, diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, level Level) Changes {
	result := make(Changes, 0)

	if diffReport == nil {
		return result
	}

	report := *diffReport
	diffReport = &report

	result = applyStabilityLevelPolicy(config, diffReport, result, operationsSources)

	diffReport.PathsDiff = withWebhookOperations(diffReport.PathsDiff, diffReport.WebhooksDiff)

	for _, check := range config.Checks {
		if check == nil {
			continue
		}
		errs := check(diffReport, operationsSources, config)
		result = append(result, errs...)
	}

	// First, so a change a transition explains counts toward nothing below.
	result = dropClaimed(result)
	result = attachSharedSchemas(result)

	// Runs before anything that reads a change's level.
	result = attachNegations(result)
	result = levelByGuards(config, result)
	result = capByDisclaimers(config, result)

	filteredResult := make(Changes, 0)
	for _, change := range result {
		if change.GetLevel() >= level {
			filteredResult = append(filteredResult, change)
		}
	}

	slices.SortFunc(filteredResult, CompareChanges)
	return filteredResult
}

const webhookPathPrefix = "webhook:"

// withWebhookOperations returns a copy of pathsDiff whose Modified also holds the modified webhooks, with a "webhook:" prefix.
// This allows all existing path/operation checker rules to apply to webhooks without code duplication.
// Only Modified webhooks are merged because each entry in Modified is self-contained with its own Base/Revision data.
// Added/Deleted webhooks cannot be merged because PathsDiff.Added/Deleted are just path name strings that require
// lookup in PathsDiff.Base/Revision (openapi3.Paths objects that don't contain webhooks).
// Added/Deleted webhooks are handled separately by WebhookUpdatedCheck.
func withWebhookOperations(pathsDiff *diff.PathsDiff, webhooksDiff *diff.WebhooksDiff) *diff.PathsDiff {
	if webhooksDiff == nil || len(webhooksDiff.Modified) == 0 {
		return pathsDiff
	}
	result := diff.PathsDiff{}
	if pathsDiff != nil {
		result = *pathsDiff
	}
	result.Modified = make(diff.ModifiedPaths, len(result.Modified)+len(webhooksDiff.Modified))
	if pathsDiff != nil {
		maps.Copy(result.Modified, pathsDiff.Modified)
	}
	for name, pathDiff := range webhooksDiff.Modified {
		result.Modified[webhookPathPrefix+name] = pathDiff
	}
	return &result
}
