package checker

import (
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/checker/internal/schemawalk"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertySunsetDeletedId             = "request-property-sunset-deleted"
	RequestPropertySunsetDateChangedEarlierId  = "request-property-sunset-date-changed-earlier"
	RequestPropertySunsetDateChangedTooSmallId = "request-property-sunset-date-changed-too-small"

	ResponsePropertySunsetDeletedId             = "response-property-sunset-deleted"
	ResponsePropertySunsetDateChangedEarlierId  = "response-property-sunset-date-changed-earlier"
	ResponsePropertySunsetDateChangedTooSmallId = "response-property-sunset-date-changed-too-small"
)

var requestPropertySunsetChangeIds = deprecation.SunsetChangeIds{
	Deleted:               RequestPropertySunsetDeletedId,
	SunsetParse:           RequestPropertyDeprecatedInvalidId,
	ChangedEarlier:        RequestPropertySunsetDateChangedEarlierId,
	ChangedEarlierComment: SunsetDateChangedEarlierCommentId,
	ChangedTooSmall:       RequestPropertySunsetDateChangedTooSmallId,
}

var responsePropertySunsetChangeIds = deprecation.SunsetChangeIds{
	Deleted:               ResponsePropertySunsetDeletedId,
	SunsetParse:           ResponsePropertyDeprecatedInvalidId,
	ChangedEarlier:        ResponsePropertySunsetDateChangedEarlierId,
	ChangedEarlierComment: SunsetDateChangedEarlierCommentId,
	ChangedTooSmall:       ResponsePropertySunsetDateChangedTooSmallId,
}

// RequestPropertySunsetChangedCheck detects changed sunset dates of deprecated properties in request bodies
func RequestPropertySunsetChangedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	return propertySunsetChangedCheck(walkModifiedRequestBodySchemas, diffReport, operationsSources, config, requestPropertySunsetChangeIds)
}

// ResponsePropertySunsetChangedCheck detects changed sunset dates of deprecated properties in response bodies
func ResponsePropertySunsetChangedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	return propertySunsetChangedCheck(walkModifiedResponseSchemas, diffReport, operationsSources, config, responsePropertySunsetChangeIds)
}

// propertySunsetChangedCheck reports a changed sunset date on each property
// the walker reaches that stays deprecated. The ids decide which side it
// reports, request or response, so the two checks differ only in their walker
// and ids.
func propertySunsetChangedCheck(
	walk func(*diff.Diff, diff.OperationsSourcesMap, *Config, func(mediaTypeInfo)),
	diffReport *diff.Diff,
	operationsSources diff.OperationsSourcesMap,
	config *Config,
	ids deprecation.SunsetChangeIds,
) Changes {
	result := make(Changes, 0)

	walk(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		// an unparseable stability level is reported in CheckBackwardCompatibility,
		// and only the grace period depends on it
		days, daysKnown := deprecationDays(config, info.operationItem.Revision.Extensions)

		info.walkProperties(func(p propertyInfo) {
			// a property deprecated or reactivated in this change is reported by
			// propertyDeprecationCheck
			stillDeprecated := p.propertyDiff.Revision.Deprecated && p.propertyDiff.DeprecatedDiff == nil
			if !stillDeprecated {
				return
			}

			id, extra, comment, ok := deprecation.SunsetChange(p.propertyDiff.ExtensionsDiff, p.propertyDiff.Base.Extensions, p.propertyDiff.Revision.Extensions, days, daysKnown, ids)
			if !ok {
				return
			}
			args := append([]any{schemawalk.PropertyFullName(p.propertyPath, p.propertyName)}, extra...)

			baseSource, revisionSource := location.SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, diff.SunsetExtension)
			if id == ids.SunsetParse {
				// the revision's value is the one that failed to parse, so the base
				// is not part of this finding
				baseSource = nil
			}
			result = append(result, p.newChange(id, args, comment).WithSources(baseSource, revisionSource))
		})
	})

	return result
}
