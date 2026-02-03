package checker

import (
	"fmt"
	"strings"
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

// formatDeprecationDetails formats optional deprecation details (stability level only)
// Returns empty string if stability is not set or invalid, otherwise returns formatted string like "(stability: X)"
func formatDeprecationDetails(extensions map[string]any) string {
	stability, err := getStabilityLevel(extensions)
	if err != nil || stability == "" {
		return ""
	}
	return "(stability: " + stability + ")"
}

// formatDeprecationDetailsWithSunset formats deprecation details with sunset date and optional stability level
// Returns formatted string like "(sunset: X)" or "(sunset: X, stability: Y)"
func formatDeprecationDetailsWithSunset(sunset civil.Date, extensions map[string]any) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("sunset: %s", sunset.String()))

	stability, err := getStabilityLevel(extensions)
	if err == nil && stability != "" {
		parts = append(parts, fmt.Sprintf("stability: %s", stability))
	}

	return "(" + strings.Join(parts, ", ") + ")"
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
				result = append(result, NewApiChange(
					EndpointReactivatedId,
					config,
					nil,
					"",
					operationsSources,
					op,
					operation,
					path,
				).WithDetails(formatDeprecationDetails(op.Extensions)))
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
					result = append(result, NewApiChange(
						EndpointDeprecatedId,
						config,
						nil,
						"",
						operationsSources,
						op,
						operation,
						path,
					).WithDetails(formatDeprecationDetails(op.Extensions)))
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
			result = append(result, NewApiChange(
				EndpointDeprecatedId,
				config,
				nil,
				"",
				operationsSources,
				op,
				operation,
				path,
			).WithDetails(formatDeprecationDetailsWithSunset(date, op.Extensions)))
		}
	}

	return result
}
