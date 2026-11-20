package checker

import (
	"fmt"
	"strings"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	EndpointReactivatedId          = "endpoint-reactivated"
	APIDeprecatedSunsetParseId     = "api-deprecated-sunset-parse"
	APIDeprecatedSunsetMissingId   = "api-deprecated-sunset-missing"
	APIInvalidStabilityLevelId     = "api-invalid-stability-level"
	APISunsetDateTooSmallId        = "api-sunset-date-too-small"
	EndpointDeprecatedId           = "endpoint-deprecated"
	EndpointDeprecatedWithSunsetId = "endpoint-deprecated-with-sunset"
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

var apiDeprecationIds = deprecation.Ids{
	Reactivated:    EndpointReactivatedId,
	Deprecated:     EndpointDeprecatedId,
	WithSunset:     EndpointDeprecatedWithSunsetId,
	SunsetMissing:  APIDeprecatedSunsetMissingId,
	SunsetParse:    APIDeprecatedSunsetParseId,
	SunsetTooSmall: APISunsetDateTooSmallId,
}

func APIDeprecationCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathItem.OperationsDiff.Modified {
			if operationDiff.DeprecatedDiff == nil {
				continue
			}

			op := pathItem.Revision.GetOperation(operation)

			days, ok := deprecationDays(config, op.Extensions)
			if !ok && deprecation.NowDeprecated(operationDiff.DeprecatedDiff) {
				// handled in CheckBackwardCompatibility
				continue
			}

			stage := deprecation.Announce(operationDiff.DeprecatedDiff, op.Extensions, days, apiDeprecationIds)

			field := "deprecated"
			if stage.AtSunset {
				// A finding about the sunset date points at x-sunset, not at deprecated.
				field = diff.SunsetExtension
			}
			baseSource, revisionSource := location.OperationFieldSources(operationsSources, operationDiff, field)

			// The details describe the deprecation, so they are left off a
			// finding about the sunset date.
			details := ""
			switch stage.Id {
			case EndpointReactivatedId, EndpointDeprecatedId, EndpointDeprecatedWithSunsetId:
				details = formatDeprecationDetails(op.Extensions)
			}

			result = append(result, newOpInfo(config, op, operationsSources, operation, path).NewApiChange(stage.Id, stage.Args, "").
				WithSources(baseSource, revisionSource).WithDetails(details))
		}
	}

	return result
}
