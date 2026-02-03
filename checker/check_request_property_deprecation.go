package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestPropertyDeprecatedId              = "request-property-deprecated"
	RequestPropertyDeprecatedSunsetMissingId = "request-property-deprecated-sunset-missing"
	RequestPropertyDeprecatedInvalidId       = "request-property-deprecated-sunset-invalid"
	RequestPropertyReactivatedId             = "request-property-reactivated"
	RequestPropertySunsetDateTooSmallId      = "request-property-sunset-date-too-small"
)

// RequestPropertyDeprecationCheck detects deprecated properties in request bodies
func RequestPropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.RequestBodyDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff == nil ||
				operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified == nil {
				continue
			}

			op := operationItem.Revision

			stability, err := getStabilityLevel(op.Extensions)
			if err != nil {
				// handled in CheckBackwardCompatibility
				continue
			}

			deprecationDays := getDeprecationDays(config, stability)

			modifiedMediaTypes := operationItem.RequestBodyDiff.ContentDiff.MediaTypeModified
			for mediaType, mediaTypeDiff := range modifiedMediaTypes {
				mediaTypeDetails := formatMediaTypeDetails(mediaType, len(modifiedMediaTypes))
				CheckModifiedPropertiesDiff(
					mediaTypeDiff.SchemaDiff,
					func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
						// Check if deprecation status changed
						if propertyDiff.DeprecatedDiff == nil {
							return
						}

						propName := propertyFullName(propertyPath, propertyName)

						// Check if property was reactivated (un-deprecated)
						if propertyDiff.DeprecatedDiff.To == nil || propertyDiff.DeprecatedDiff.To == false {
							result = append(result, NewApiChange(
								RequestPropertyReactivatedId,
								config,
								[]any{propName},
								"",
								operationsSources,
								op,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
							return
						}

						// Property was newly deprecated (To=true means deprecated now)
						sunset, ok := getSunset(propertyDiff.Revision.Extensions)
						if !ok {
							// if deprecation policy is defined and sunset is missing, it's a breaking change
							if deprecationDays > 0 {
								result = append(result, NewApiChange(
									RequestPropertyDeprecatedSunsetMissingId,
									config,
									[]any{propName},
									"",
									operationsSources,
									op,
									operation,
									path,
								).WithDetails(mediaTypeDetails))
							} else {
								// no policy, report deprecation without sunset as INFO
								result = append(result, NewApiChange(
									RequestPropertyDeprecatedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									op,
									operation,
									path,
								).WithDetails(combineDetails(formatDeprecationDetails(op.Extensions), mediaTypeDetails)))
							}
							return
						}

						date, err := getSunsetDate(sunset)
						if err != nil {
							result = append(result, NewApiChange(
								RequestPropertyDeprecatedInvalidId,
								config,
								[]any{propName, err},
								"",
								operationsSources,
								op,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
							return
						}

						days := date.DaysSince(civil.DateOf(time.Now()))

						if days < int(deprecationDays) {
							result = append(result, NewApiChange(
								RequestPropertySunsetDateTooSmallId,
								config,
								[]any{propName, date, deprecationDays},
								"",
								operationsSources,
								op,
								operation,
								path,
							).WithDetails(mediaTypeDetails))
							return
						}

						// not breaking changes
						result = append(result, NewApiChange(
							RequestPropertyDeprecatedId,
							config,
							[]any{propName},
							"",
							operationsSources,
							op,
							operation,
							path,
						).WithDetails(combineDetails(formatDeprecationDetailsWithSunset(date, op.Extensions), mediaTypeDetails)))
					})
			}
		}
	}
	return result
}
