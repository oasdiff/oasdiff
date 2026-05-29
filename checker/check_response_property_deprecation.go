package checker

import (
	"time"

	"cloud.google.com/go/civil"
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

// ResponsePropertyDeprecationCheck detects deprecated properties in response bodies
func ResponsePropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedResponseSchemas(diffReport, operationsSources, config, func(info mediaTypeInfo) {
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
					ResponsePropertyReactivatedId,
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
						ResponsePropertyDeprecatedSunsetMissingId,
						[]any{propName},
						"",
					).WithSources(propBaseSource, propRevisionSource))
				} else {
					// no policy, report deprecation without sunset as INFO
					result = append(result, p.newChange(
						ResponsePropertyDeprecatedId,
						[]any{propName},
						"",
					).WithSources(propBaseSource, propRevisionSource).WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails)))
				}
				return
			}

			date, err := getSunsetDate(sunset)
			if err != nil {
				result = append(result, p.newChange(
					ResponsePropertyDeprecatedInvalidId,
					[]any{propName, err},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}

			days := date.DaysSince(civil.DateOf(time.Now()))

			if days < int(deprecationDays) {
				result = append(result, p.newChange(
					ResponsePropertySunsetDateTooSmallId,
					[]any{propName, date, deprecationDays},
					"",
				).WithSources(propBaseSource, propRevisionSource))
				return
			}

			// not breaking changes
			result = append(result, p.newChange(
				ResponsePropertyDeprecatedWithSunsetId,
				[]any{propName, date},
				"",
			).WithSources(propBaseSource, propRevisionSource).WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), info.mediaTypeDetails)))
		})
	})

	return result
}
