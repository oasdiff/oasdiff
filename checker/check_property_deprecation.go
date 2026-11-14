package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyDeprecatedId              = "request-property-deprecated"
	RequestPropertyDeprecatedWithSunsetId    = "request-property-deprecated-with-sunset"
	RequestPropertyDeprecatedSunsetMissingId = "request-property-deprecated-sunset-missing"
	RequestPropertyDeprecatedInvalidId       = "request-property-deprecated-sunset-invalid"
	RequestPropertyReactivatedId             = "request-property-reactivated"
	RequestPropertySunsetDateTooSmallId      = "request-property-sunset-date-too-small"

	ResponsePropertyDeprecatedId              = "response-property-deprecated"
	ResponsePropertyDeprecatedWithSunsetId    = "response-property-deprecated-with-sunset"
	ResponsePropertyDeprecatedSunsetMissingId = "response-property-deprecated-sunset-missing"
	ResponsePropertyDeprecatedInvalidId       = "response-property-deprecated-sunset-invalid"
	ResponsePropertyReactivatedId             = "response-property-reactivated"
	ResponsePropertySunsetDateTooSmallId      = "response-property-sunset-date-too-small"
)

var requestPropertyDeprecationIds = deprecation.Ids{
	Reactivated:    RequestPropertyReactivatedId,
	Deprecated:     RequestPropertyDeprecatedId,
	WithSunset:     RequestPropertyDeprecatedWithSunsetId,
	SunsetMissing:  RequestPropertyDeprecatedSunsetMissingId,
	SunsetParse:    RequestPropertyDeprecatedInvalidId,
	SunsetTooSmall: RequestPropertySunsetDateTooSmallId,
}

var responsePropertyDeprecationIds = deprecation.Ids{
	Reactivated:    ResponsePropertyReactivatedId,
	Deprecated:     ResponsePropertyDeprecatedId,
	WithSunset:     ResponsePropertyDeprecatedWithSunsetId,
	SunsetMissing:  ResponsePropertyDeprecatedSunsetMissingId,
	SunsetParse:    ResponsePropertyDeprecatedInvalidId,
	SunsetTooSmall: ResponsePropertySunsetDateTooSmallId,
}

// RequestPropertyDeprecationCheck detects deprecated properties in request bodies
func RequestPropertyDeprecationCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	return propertyDeprecationCheck(walkModifiedRequestBodySchemas, diffReport, operationsSources, config, requestPropertyDeprecationIds)
}

// ResponsePropertyDeprecationCheck detects deprecated properties in response bodies
func ResponsePropertyDeprecationCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	return propertyDeprecationCheck(walkModifiedResponseSchemas, diffReport, operationsSources, config, responsePropertyDeprecationIds)
}

// propertyDeprecationCheck reports where each property the walker reaches
// stands in the deprecation process. The ids decide which side it reports,
// request or response, so the two checks differ only in their walker and ids.
func propertyDeprecationCheck(
	walk func(*diff.Diff, diff.OperationsSourcesMap, *Config, func(mediaTypeInfo)),
	diffReport *diff.Diff,
	operationsSources diff.OperationsSourcesMap,
	config *Config,
	ids deprecation.Ids,
) Changes {
	result := make(Changes, 0)

	walk(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		op := info.operationItem.Revision

		days, ok := deprecationDays(config, op.Extensions)
		if !ok {
			// handled in CheckBackwardCompatibility
			return
		}

		info.walkProperties(func(p propertyInfo) {
			if p.propertyDiff.DeprecatedDiff == nil {
				return
			}

			stage := deprecation.Announce(p.propertyDiff.DeprecatedDiff, p.propertyDiff.Revision.Extensions, days, ids)

			field := "deprecated"
			if stage.AtSunset {
				// A finding about the sunset date points at x-sunset, not at deprecated.
				field = diff.SunsetExtension
			}
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, field)

			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			change := p.newChange(stage.Id, append([]any{propName}, stage.Args...), "").
				WithSources(baseSource, revisionSource)

			// The stability level describes the deprecation, so it is left off
			// a finding about the sunset date, and off a reactivation. The
			// media type stays on every change, as newChange put it there.
			if stage.Id == ids.Deprecated || stage.Id == ids.WithSunset {
				change = change.WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails))
			}

			result = append(result, change)
		})
	})

	return result
}
