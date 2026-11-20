package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APISunsetDeletedId             = "sunset-deleted"
	APISunsetDateChangedEarlierId  = "api-sunset-date-changed-earlier"
	APISunsetDateChangedTooSmallId = "api-sunset-date-changed-too-small"
)

// SunsetDateChangedEarlierCommentId explains a verdict a provider with no
// deprecation policy would not expect: the date itself is the commitment.
const SunsetDateChangedEarlierCommentId = "sunset-date-changed-earlier-comment"

var apiSunsetChangeIds = deprecation.SunsetChangeIds{
	Deleted:               APISunsetDeletedId,
	SunsetParse:           APIPathSunsetParseId,
	ChangedEarlier:        APISunsetDateChangedEarlierId,
	ChangedEarlierComment: SunsetDateChangedEarlierCommentId,
	ChangedTooSmall:       APISunsetDateChangedTooSmallId,
}

func APISunsetChangedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathItem.OperationsDiff.Modified {
			opRevision := pathItem.Revision.GetOperation(operation)
			if !opRevision.Deprecated {
				continue
			}

			// an unparseable stability level is reported in CheckBackwardCompatibility,
			// and only the grace period depends on it
			days, daysKnown := deprecationDays(config, opRevision.Extensions)
			opBase := pathItem.Base.GetOperation(operation)

			id, args, comment, ok := deprecation.SunsetChange(operationDiff.ExtensionsDiff, opBase.Extensions, opRevision.Extensions, days, daysKnown, apiSunsetChangeIds)
			if !ok {
				continue
			}

			baseSource, revisionSource := location.OperationFieldSources(operationsSources, operationDiff, diff.SunsetExtension)
			if id == APIPathSunsetParseId {
				// the revision's value is the one that failed to parse, so the base
				// is not part of this finding
				baseSource = nil
			}
			result = append(result, NewApiChange(
				id,
				config,
				args,
				comment,
				operationsSources,
				opRevision,
				operation,
				path,
			).WithSources(baseSource, revisionSource))
		}
	}

	return result
}
