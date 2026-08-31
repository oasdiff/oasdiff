package checker

import (
	"slices"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterSunsetDeletedId             = "request-parameter-sunset-deleted"
	RequestParameterSunsetDateChangedTooSmallId = "request-parameter-sunset-date-changed-too-small"
)

func RequestParameterSunsetChangedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)

	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {
		baseSource, revisionSource := parameterFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff, diff.SunsetExtension)

		paramBase := p.paramDiff.Base
		paramRevision := p.paramDiff.Revision

		if !paramRevision.Deprecated {
			return
		}

		if p.paramDiff.ExtensionsDiff == nil {
			return
		}

		if slices.Contains(p.paramDiff.ExtensionsDiff.Deleted, diff.SunsetExtension) {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterSunsetDeletedId,
				[]any{p.location, p.name},
				"",
			).WithSources(baseSource, revisionSource))
			return
		}

		if _, ok := p.paramDiff.ExtensionsDiff.Modified[diff.SunsetExtension]; !ok {
			return
		}

		date, err := getSunsetDate(paramRevision.Extensions[diff.SunsetExtension])
		if err != nil {
			result = append(result, getRequestParameterSunsetParse(p.opInfo, paramRevision, err).WithSources(nil, revisionSource))
			return
		}

		baseDate, err := getSunsetDate(paramBase.Extensions[diff.SunsetExtension])
		if err != nil {
			opInfo := newOpInfo(config, p.opInfo.methodDiff.Base, operationsSources, p.opInfo.method, p.opInfo.path)
			result = append(result, getRequestParameterSunsetParse(opInfo, paramBase, err).WithSources(baseSource, nil))
			return
		}

		days := date.DaysSince(civil.DateOf(time.Now()))

		stability, err := getStabilityLevel(p.opInfo.operation.Extensions)
		if err != nil {
			// handled in CheckBackwardCompatibility
			return
		}

		deprecationDays := getDeprecationDays(config, stability)

		if baseDate.After(date) && days < int(deprecationDays) {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterSunsetDateChangedTooSmallId,
				[]any{paramRevision.In, paramRevision.Name, baseDate, date, baseDate, deprecationDays},
				"",
			).WithSources(baseSource, revisionSource))
		}
	})

	return result
}
