package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/location"
	"github.com/oasdiff/oasdiff/checker/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyDeprecatedId              = "request-property-deprecated"
	RequestPropertyDeprecatedWithSunsetId    = "request-property-deprecated-with-sunset"
	RequestPropertyDeprecatedSunsetMissingId = "request-property-deprecated-sunset-missing"
	RequestPropertyDeprecatedInvalidId       = "request-property-deprecated-sunset-invalid"
	RequestPropertyReactivatedId             = "request-property-reactivated"
	RequestPropertySunsetDateTooSmallId      = "request-property-sunset-date-too-small"
)

var requestPropertyDeprecationIds = deprecation.Ids{
	Reactivated:    RequestPropertyReactivatedId,
	Deprecated:     RequestPropertyDeprecatedId,
	WithSunset:     RequestPropertyDeprecatedWithSunsetId,
	SunsetMissing:  RequestPropertyDeprecatedSunsetMissingId,
	SunsetParse:    RequestPropertyDeprecatedInvalidId,
	SunsetTooSmall: RequestPropertySunsetDateTooSmallId,
}

// RequestPropertyDeprecationCheck detects deprecated properties in request bodies
func RequestPropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
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

			stage := deprecation.Announce(p.propertyDiff.DeprecatedDiff, p.propertyDiff.Revision.Extensions, days, requestPropertyDeprecationIds)

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
			if stage.Id == RequestPropertyDeprecatedId || stage.Id == RequestPropertyDeprecatedWithSunsetId {
				change = change.WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails))
			}

			result = append(result, change)
		})
	})

	return result
}
