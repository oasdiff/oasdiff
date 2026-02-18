package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	APISunsetDeletedId             = "sunset-deleted"
	APISunsetDateChangedTooSmallId = "api-sunset-date-changed-too-small"
)

func APISunsetChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathItem.OperationsDiff.Modified {
			opRevision := pathItem.Revision.GetOperation(operation)
			opBase := pathItem.Base.GetOperation(operation)
			baseSource, revisionSource := operationSources(operationsSources, operationDiff.Base, operationDiff.Revision)

			if !opRevision.Deprecated {
				continue
			}

			if operationDiff.ExtensionsDiff == nil {
				continue
			}

			if operationDiff.ExtensionsDiff.Deleted.Contains(diff.SunsetExtension) {
				result = append(result, NewApiChange(
					APISunsetDeletedId,
					config,
					nil,
					"",
					operationsSources,
					opRevision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
				continue
			}

			if _, ok := operationDiff.ExtensionsDiff.Modified[diff.SunsetExtension]; !ok {
				continue
			}

			date, err := getSunsetDate(opRevision.Extensions[diff.SunsetExtension])
			if err != nil {
				opInfo := newOpInfo(config, opRevision, operationsSources, operation, path)
				result = append(result, getAPIPathSunsetParse(opInfo, err).WithSources(baseSource, revisionSource))
				continue
			}

			baseDate, err := getSunsetDate(opBase.Extensions[diff.SunsetExtension])
			if err != nil {
				opInfo := newOpInfo(config, opBase, operationsSources, operation, path)
				result = append(result, getAPIPathSunsetParse(opInfo, err).WithSources(baseSource, revisionSource))
				continue
			}

			days := date.DaysSince(civil.DateOf(time.Now()))

			stability, err := getStabilityLevel(opRevision.Extensions)
			if err != nil {
				// handled in CheckBackwardCompatibility
				continue
			}

			deprecationDays := getDeprecationDays(config, stability)

			if baseDate.After(date) && days < int(deprecationDays) {
				result = append(result, NewApiChange(
					APISunsetDateChangedTooSmallId,
					config,
					[]any{baseDate, date, baseDate, deprecationDays},
					"",
					operationsSources,
					opRevision,
					operation,
					path,
				).WithSources(baseSource, revisionSource))
			}
		}
	}

	return result
}

const (
	STABILITY_DRAFT  = "draft"
	STABILITY_ALPHA  = "alpha"
	STABILITY_BETA   = "beta"
	STABILITY_STABLE = "stable"
)

func getDeprecationDays(config *Config, stability string) uint {
	switch stability {
	case STABILITY_DRAFT:
		return 0
	case STABILITY_ALPHA:
		return 0
	case STABILITY_BETA:
		return config.MinSunsetBetaDays
	case STABILITY_STABLE:
		return config.MinSunsetStableDays
	default:
		return config.MinSunsetStableDays
	}
}

func getAPIDeprecatedSunsetMissing(opInfo opInfo) ApiChange {
	return NewApiChange(
		APIDeprecatedSunsetMissingId,
		opInfo.config,
		nil,
		"",
		opInfo.operationsSources,
		opInfo.operation,
		opInfo.method,
		opInfo.path,
	)
}
