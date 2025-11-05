package checker

import (
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	EndpointReactivatedId        = "endpoint-reactivated"
	APIDeprecatedSunsetParseId   = "api-deprecated-sunset-parse"
	APIDeprecatedSunsetMissingId = "api-deprecated-sunset-missing"
	APIInvalidStabilityLevelId   = "api-invalid-stability-level"
	APISunsetDateTooSmallId      = "api-sunset-date-too-small"
	EndpointDeprecatedId         = "endpoint-deprecated"
)

// formatDeprecationDetails formats optional deprecation details (sunset date and stability level)
// Returns empty string if neither is set, otherwise returns formatted string like " (sunset: X, stability: Y)"
func formatDeprecationDetails(sunset interface{}, stability string) string {
	var parts []string

	sunsetStr := ""
	if sunset != nil {
		if s, ok := sunset.(string); ok && s != "" {
			sunsetStr = s
		} else if date, ok := sunset.(civil.Date); ok {
			sunsetStr = date.String()
		}
	}

	if sunsetStr != "" {
		parts = append(parts, fmt.Sprintf("sunset: %s", sunsetStr))
	}

	if stability != "" {
		parts = append(parts, fmt.Sprintf("stability: %s", stability))
	}

	if len(parts) == 0 {
		return ""
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	return " (" + result + ")"
}

func APIDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathItem.OperationsDiff.Modified {
			op := pathItem.Revision.GetOperation(operation)

			if operationDiff.DeprecatedDiff == nil {
				continue
			}

			if operationDiff.DeprecatedDiff.To == nil || operationDiff.DeprecatedDiff.To == false {
				// not breaking changes
				stability, _ := getStabilityLevel(op.Extensions)
				details := formatDeprecationDetails(nil, stability)
				result = append(result, NewApiChange(
					EndpointReactivatedId,
					config,
					nil,
					details,
					operationsSources,
					op,
					operation,
					path,
				))
				continue
			}

			stability, err := getStabilityLevel(op.Extensions)
			if err != nil {
				// handled in CheckBackwardCompatibility
				continue
			}

			deprecationDays := getDeprecationDays(config, stability)

			sunset, ok := getSunset(op.Extensions)
			if !ok {
				// if deprecation policy is defined and sunset is missing, it's a breaking change
				if deprecationDays > 0 {
					result = append(result, getAPIDeprecatedSunsetMissing(newOpInfo(config, op, operationsSources, operation, path)))
				} else {
					// no policy, report deprecation without sunset as INFO
					details := formatDeprecationDetails(nil, stability)
					result = append(result, NewApiChange(
						EndpointDeprecatedId,
						config,
						nil,
						details,
						operationsSources,
						op,
						operation,
						path,
					))
				}
				continue
			}

			date, err := getSunsetDate(sunset)
			if err != nil {
				result = append(result, NewApiChange(
					APIDeprecatedSunsetParseId,
					config,
					[]any{err},
					"",
					operationsSources,
					op,
					operation,
					path,
				))
				continue
			}

			days := date.DaysSince(civil.DateOf(time.Now()))

			if days < int(deprecationDays) {
				result = append(result, NewApiChange(
					APISunsetDateTooSmallId,
					config,
					[]any{date, deprecationDays},
					"",
					operationsSources,
					op,
					operation,
					path,
				))
				continue
			}

			// not breaking changes
			details := formatDeprecationDetails(date, stability)
			result = append(result, NewApiChange(
				EndpointDeprecatedId,
				config,
				nil,
				details,
				operationsSources,
				op,
				operation,
				path,
			))
		}
	}

	return result
}
