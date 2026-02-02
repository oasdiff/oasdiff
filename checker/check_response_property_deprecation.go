package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	ResponsePropertyDeprecatedId              = "response-property-deprecated"
	ResponsePropertyDeprecatedSunsetMissingId = "response-property-deprecated-sunset-missing"
	ResponsePropertyDeprecatedInvalidId       = "response-property-deprecated-sunset-invalid"
	ResponsePropertyReactivatedId             = "response-property-reactivated"
	ResponsePropertySunsetDateTooSmallId      = "response-property-sunset-date-too-small"
)

// ResponsePropertyDeprecationCheck detects deprecated properties in response bodies
func ResponsePropertyDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ResponsesDiff == nil || operationItem.ResponsesDiff.Modified == nil {
				continue
			}

			op := operationItem.Revision

			stability, err := getStabilityLevel(op.Extensions)
			if err != nil {
				// handled in CheckBackwardCompatibility
				continue
			}

			deprecationDays := getDeprecationDays(config, stability)

			// Track reported properties to avoid duplicates per operation
			reportedProperties := make(map[string]bool)

			for _, responseDiff := range operationItem.ResponsesDiff.Modified {
				if responseDiff == nil ||
					responseDiff.ContentDiff == nil ||
					responseDiff.ContentDiff.MediaTypeModified == nil {
					continue
				}
				modifiedMediaTypes := responseDiff.ContentDiff.MediaTypeModified
				for _, mediaTypeDiff := range modifiedMediaTypes {
					CheckModifiedPropertiesDiff(
						mediaTypeDiff.SchemaDiff,
						func(propertyPath string, propertyName string, propertyDiff *diff.SchemaDiff, parent *diff.SchemaDiff) {
							// Check if deprecation status changed
							if propertyDiff.DeprecatedDiff == nil {
								return
							}

							propName := propertyFullName(propertyPath, propertyName)

							// Skip if already reported
							if reportedProperties[propName] {
								return
							}
							reportedProperties[propName] = true

							// Check if property was reactivated (un-deprecated)
							if propertyDiff.DeprecatedDiff.To == nil || propertyDiff.DeprecatedDiff.To == false {
								result = append(result, NewApiChange(
									ResponsePropertyReactivatedId,
									config,
									[]any{propName},
									"",
									operationsSources,
									op,
									operation,
									path,
								))
								return
							}

							// Property was newly deprecated (To=true means deprecated now)
							sunset, ok := getSunset(propertyDiff.Revision.Extensions)
							if !ok {
								// if deprecation policy is defined and sunset is missing, it's a breaking change
								if deprecationDays > 0 {
									result = append(result, NewApiChange(
										ResponsePropertyDeprecatedSunsetMissingId,
										config,
										[]any{propName},
										"",
										operationsSources,
										op,
										operation,
										path,
									))
								}
								return
							}

							date, err := getSunsetDate(sunset)
							if err != nil {
								result = append(result, NewApiChange(
									ResponsePropertyDeprecatedInvalidId,
									config,
									[]any{propName, err},
									"",
									operationsSources,
									op,
									operation,
									path,
								))
								return
							}

							days := date.DaysSince(civil.DateOf(time.Now()))

							if days < int(deprecationDays) {
								result = append(result, NewApiChange(
									ResponsePropertySunsetDateTooSmallId,
									config,
									[]any{propName, date, deprecationDays},
									"",
									operationsSources,
									op,
									operation,
									path,
								))
								return
							}

							// not breaking changes
							result = append(result, NewApiChange(
								ResponsePropertyDeprecatedId,
								config,
								[]any{propName, date},
								"",
								operationsSources,
								op,
								operation,
								path,
							))
						})
				}
			}
		}
	}
	return result
}
