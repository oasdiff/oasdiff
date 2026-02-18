package checker

import (
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
			baseSource, revisionSource := operationSources(operationsSources, operationDiff.Base, operationDiff.Revision)

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
				).WithSources(baseSource, revisionSource))
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
					result = append(result, getAPIDeprecatedSunsetMissing(newOpInfo(config, op, operationsSources, operation, path)).WithSources(baseSource, revisionSource))
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
				).WithSources(baseSource, revisionSource))
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
				).WithSources(baseSource, revisionSource))
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
			).WithSources(baseSource, revisionSource))
		}
	}

	return result
}
