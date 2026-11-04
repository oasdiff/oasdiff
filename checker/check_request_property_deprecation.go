package checker

import (
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

var requestPropertyDeprecationIds = deprecationIds{
	reactivated:    RequestPropertyReactivatedId,
	deprecated:     RequestPropertyDeprecatedId,
	withSunset:     RequestPropertyDeprecatedWithSunsetId,
	sunsetMissing:  RequestPropertyDeprecatedSunsetMissingId,
	sunsetParse:    RequestPropertyDeprecatedInvalidId,
	sunsetTooSmall: RequestPropertySunsetDateTooSmallId,
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

			stage := deprecationId(p.propertyDiff.DeprecatedDiff, p.propertyDiff.Revision.Extensions, days, requestPropertyDeprecationIds)

			field := "deprecated"
			if stage.atSunset {
				// A finding about the sunset date points at x-sunset, not at deprecated.
				field = diff.SunsetExtension
			}
			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, field)

			propName := schemawalk.PropertyFullName(p.propertyPath, p.propertyName)
			change := p.newChange(stage.id, append([]any{propName}, stage.args...), "").
				WithSources(baseSource, revisionSource)

			// The stability level describes the deprecation, so it is left off
			// a finding about the sunset date, and off a reactivation. The
			// media type stays on every change, as newChange put it there.
			if stage.id == RequestPropertyDeprecatedId || stage.id == RequestPropertyDeprecatedWithSunsetId {
				change = change.WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails))
			}

			result = append(result, change)
		})
	})

	return result
}
