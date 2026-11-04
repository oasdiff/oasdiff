package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyDeprecatedId              = "response-property-deprecated"
	ResponsePropertyDeprecatedWithSunsetId    = "response-property-deprecated-with-sunset"
	ResponsePropertyDeprecatedSunsetMissingId = "response-property-deprecated-sunset-missing"
	ResponsePropertyDeprecatedInvalidId       = "response-property-deprecated-sunset-invalid"
	ResponsePropertyReactivatedId             = "response-property-reactivated"
	ResponsePropertySunsetDateTooSmallId      = "response-property-sunset-date-too-small"
)

var responsePropertyDeprecationIds = deprecation.Ids{
	Reactivated:    ResponsePropertyReactivatedId,
	Deprecated:     ResponsePropertyDeprecatedId,
	WithSunset:     ResponsePropertyDeprecatedWithSunsetId,
	SunsetMissing:  ResponsePropertyDeprecatedSunsetMissingId,
	SunsetParse:    ResponsePropertyDeprecatedInvalidId,
	SunsetTooSmall: ResponsePropertySunsetDateTooSmallId,
}

// ResponsePropertyDeprecationCheck detects deprecated properties in response bodies
func ResponsePropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
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

			stage := deprecation.Announce(p.propertyDiff.DeprecatedDiff, p.propertyDiff.Revision.Extensions, days, responsePropertyDeprecationIds)

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
			if stage.Id == ResponsePropertyDeprecatedId || stage.Id == ResponsePropertyDeprecatedWithSunsetId {
				change = change.WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails))
			}

			result = append(result, change)
		})
	})

	return result
}
