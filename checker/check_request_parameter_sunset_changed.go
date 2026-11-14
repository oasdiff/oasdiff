package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterSunsetDeletedId             = "request-parameter-sunset-deleted"
	RequestParameterSunsetDateChangedEarlierId  = "request-parameter-sunset-date-changed-earlier"
	RequestParameterSunsetDateChangedTooSmallId = "request-parameter-sunset-date-changed-too-small"
)

var requestParameterSunsetChangeIds = deprecation.SunsetChangeIds{
	Deleted:               RequestParameterSunsetDeletedId,
	SunsetParse:           RequestParameterSunsetParseId,
	ChangedEarlier:        RequestParameterSunsetDateChangedEarlierId,
	ChangedEarlierComment: SunsetDateChangedEarlierCommentId,
	ChangedTooSmall:       RequestParameterSunsetDateChangedTooSmallId,
}

func RequestParameterSunsetChangedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if !p.paramDiff.Revision.Deprecated {
			return
		}

		// an unparseable stability level is reported in CheckBackwardCompatibility,
		// and only the grace period depends on it
		days, daysKnown := deprecationDays(config, p.opInfo.operation.Extensions)

		id, extra, comment, ok := deprecation.SunsetChange(p.paramDiff.ExtensionsDiff, p.paramDiff.Base.Extensions, p.paramDiff.Revision.Extensions, days, daysKnown, requestParameterSunsetChangeIds)
		if !ok {
			return
		}
		args := append([]any{p.location, p.name}, extra...)

		baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, diff.SunsetExtension)
		if id == RequestParameterSunsetParseId {
			// the revision's value is the one that failed to parse, so the base
			// is not part of this finding
			baseSource = nil
		}
		result = append(result, p.opInfo.NewApiChange(id, args, comment).WithSources(baseSource, revisionSource))
	})

	return result
}
