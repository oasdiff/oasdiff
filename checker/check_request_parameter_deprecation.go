package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterReactivatedId             = "request-parameter-reactivated"
	RequestParameterDeprecatedSunsetMissingId = "request-parameter-deprecated-sunset-missing"
	RequestParameterSunsetDateTooSmallId      = "request-parameter-sunset-date-too-small"
	RequestParameterDeprecatedId              = "request-parameter-deprecated"
)

// withSunset is empty: a parameter deprecated with a sunset date is reported as
// request-parameter-deprecated, with the date in the details.
var requestParameterDeprecationIds = deprecation.Ids{
	Reactivated:    RequestParameterReactivatedId,
	Deprecated:     RequestParameterDeprecatedId,
	SunsetMissing:  RequestParameterDeprecatedSunsetMissingId,
	SunsetParse:    RequestParameterSunsetParseId,
	SunsetTooSmall: RequestParameterSunsetDateTooSmallId,
}

func RequestParameterDeprecationCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.DeprecatedDiff == nil {
			return
		}

		days, ok := deprecationDays(config, p.opInfo.operation.Extensions)
		if !ok && deprecation.NowDeprecated(p.paramDiff.DeprecatedDiff) {
			// handled in CheckBackwardCompatibility
			return
		}

		stage := deprecation.Announce(p.paramDiff.DeprecatedDiff, p.paramDiff.Revision.Extensions, days, requestParameterDeprecationIds)

		// Every stage points at the deprecated field, the sunset ones included.
		// deprecated is a field of the Parameter Object, not of its schema.
		baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "deprecated")

		// The details describe the deprecation, so they are left off a finding
		// about the sunset date.
		details := ""
		switch stage.Id {
		case RequestParameterReactivatedId:
			details = formatDeprecationDetails(p.opInfo.operation.Extensions)
		case RequestParameterDeprecatedId:
			if stage.Sunset.IsValid() {
				details = formatDeprecationDetailsWithSunset(stage.Sunset, p.opInfo.operation.Extensions)
			} else {
				details = formatDeprecationDetails(p.opInfo.operation.Extensions)
			}
		}

		result = append(result, p.opInfo.NewApiChange(stage.Id, append([]any{p.location, p.name}, stage.Args...), "").
			WithSources(baseSource, revisionSource).WithDetails(details))
	})

	return result
}
