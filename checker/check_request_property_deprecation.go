package checker

import (
	"time"

	"cloud.google.com/go/civil"
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

// RequestPropertyDeprecationCheck detects deprecated properties in request bodies
func RequestPropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedRequestBodySchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
		op := info.operationItem.Revision

		stability, err := getStabilityLevel(op.Extensions)
		if err != nil {
			// handled in CheckBackwardCompatibility
			return
		}

		deprecationDays := getDeprecationDays(config, stability)

		info.walkProperties(func(p propertyInfo) {
			// Check if deprecation status changed
			if p.propertyDiff.DeprecatedDiff == nil {
				return
			}

			propName := propertyFullName(p.propertyPath, p.propertyName)
			propBaseSource, propRevisionSource := SchemaFieldSources(operationsSources, info.operationItem, p.propertyDiff, "deprecated")

			// Check if property was reactivated (un-deprecated)
			if p.propertyDiff.DeprecatedDiff.To == nil || p.propertyDiff.DeprecatedDiff.To == false {
				result = append(result, p.newChange(
					RequestPropertyReactivatedId,
					[]any{propName},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}

			// Property was newly deprecated (To=true means deprecated now)
			sunset, ok := getSunset(p.propertyDiff.Revision.Extensions)
			if !ok {
				// if deprecation policy is defined and sunset is missing, it's a breaking change
				if deprecationDays > 0 {
					result = append(result, p.newChange(
						RequestPropertyDeprecatedSunsetMissingId,
						[]any{propName},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					// no policy, report deprecation without sunset as INFO
					result = append(result, p.newChange(
						RequestPropertyDeprecatedId,
						[]any{propName},
						"",
					).WithSources(propBaseSource, propRevisionSource).WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails)))
				}
				return
			}

			date, err := getSunsetDate(sunset)
			if err != nil {
				result = append(result, p.newChange(
					RequestPropertyDeprecatedInvalidId,
					[]any{propName, err},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}

			days := date.DaysSince(civil.DateOf(time.Now()))

			if days < int(deprecationDays) {
				result = append(result, p.newChange(
					RequestPropertySunsetDateTooSmallId,
					[]any{propName, date, deprecationDays},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}

			// Report property deprecated with sunset date (always show, regardless of policy)
			result = append(result, p.newChange(
				RequestPropertyDeprecatedWithSunsetId,
				[]any{propName, date},
				"",
			).WithSources(propBaseSource, propRevisionSource).WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails)))
		})
	})

	return result
}
