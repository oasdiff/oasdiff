package checker

import (
	"github.com/oasdiff/oasdiff/checker/location"
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
var requestParameterDeprecationIds = deprecationIds{
	reactivated:    RequestParameterReactivatedId,
	deprecated:     RequestParameterDeprecatedId,
	sunsetMissing:  RequestParameterDeprecatedSunsetMissingId,
	sunsetParse:    RequestParameterSunsetParseId,
	sunsetTooSmall: RequestParameterSunsetDateTooSmallId,
}

func RequestParameterDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		if p.paramDiff.DeprecatedDiff == nil {
			return
		}

		days, ok := deprecationDays(config, p.opInfo.operation.Extensions)
		if !ok && nowDeprecated(p.paramDiff.DeprecatedDiff) {
			// handled in CheckBackwardCompatibility
			return
		}

		stage := deprecationId(p.paramDiff.DeprecatedDiff, p.paramDiff.Revision.Extensions, days, requestParameterDeprecationIds)

		// Every stage points at the deprecated field, the sunset ones included.
		// deprecated is a field of the Parameter Object, not of its schema.
		baseSource, revisionSource := location.ParameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, "deprecated")

		// The details describe the deprecation, so they are left off a finding
		// about the sunset date.
		details := ""
		switch stage.id {
		case RequestParameterReactivatedId:
			details = formatDeprecationDetails(p.opInfo.operation.Extensions)
		case RequestParameterDeprecatedId:
			if stage.sunset.IsValid() {
				details = formatDeprecationDetailsWithSunset(stage.sunset, p.opInfo.operation.Extensions)
			} else {
				details = formatDeprecationDetails(p.opInfo.operation.Extensions)
			}
		}

		result = append(result, p.opInfo.NewApiChange(stage.id, append([]any{p.location, p.name}, stage.args...), "").
			WithSources(baseSource, revisionSource).WithDetails(details))
	})

	return result
}
