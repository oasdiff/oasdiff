package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterReactivatedId             = "request-parameter-reactivated"
	RequestParameterDeprecatedSunsetMissingId = "request-parameter-deprecated-sunset-missing"
	RequestParameterSunsetDateTooSmallId      = "request-parameter-sunset-date-too-small"
	RequestParameterDeprecatedId              = "request-parameter-deprecated"
)

func RequestParameterDeprecationCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkModifiedParameters(diffReport, operationsSources, config, func(p paramInfo) {

		if p.paramDiff.DeprecatedDiff == nil {
			return
		}
		baseSource, revisionSource := SchemaFieldSources(operationsSources, p.opInfo.methodDiff, p.paramDiff.SchemaDiff, "deprecated")

		param := p.paramDiff.Revision

		if p.paramDiff.DeprecatedDiff.To == nil || p.paramDiff.DeprecatedDiff.To == false {
			// not breaking changes
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterReactivatedId,
				[]any{p.location, p.name},
				"",
			).WithSources(baseSource, revisionSource).WithDetails(formatDeprecationDetails(p.opInfo.operation.Extensions)))
			return
		}

		stability, err := getStabilityLevel(p.opInfo.operation.Extensions)
		if err != nil {
			// handled in CheckBackwardCompatibility
			return
		}

		deprecationDays := getDeprecationDays(config, stability)

		sunset, ok := getSunset(param.Extensions)
		if !ok {
			// if deprecation policy is defined and sunset is missing, it's a breaking change
			if deprecationDays > 0 {
				result = append(result, getParameterDeprecatedSunsetMissing(p.opInfo, param).WithSources(baseSource, revisionSource))
			} else {
				// no policy, report deprecation without sunset as INFO
				result = append(result, p.opInfo.NewApiChange(
					RequestParameterDeprecatedId,
					[]any{p.location, p.name},
					"",
				).WithSources(baseSource, revisionSource).WithDetails(formatDeprecationDetails(p.opInfo.operation.Extensions)))
			}
			return
		}

		date, err := getSunsetDate(sunset)
		if err != nil {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterSunsetParseId,
				[]any{param.In, param.Name, err},
				"",
			).WithSources(baseSource, revisionSource))
			return
		}

		days := date.DaysSince(civil.DateOf(time.Now()))

		if days < int(deprecationDays) {
			result = append(result, p.opInfo.NewApiChange(
				RequestParameterSunsetDateTooSmallId,
				[]any{param.In, param.Name, date, deprecationDays},
				"",
			).WithSources(baseSource, revisionSource))
			return
		}

		// not breaking changes
		result = append(result, p.opInfo.NewApiChange(
			RequestParameterDeprecatedId,
			[]any{p.location, p.name},
			"",
		).WithSources(baseSource, revisionSource).WithDetails(formatDeprecationDetailsWithSunset(date, p.opInfo.operation.Extensions)))
	})

	return result
}

func getParameterDeprecatedSunsetMissing(opInfo opInfo, param *openapi3.Parameter) ApiChange {
	return opInfo.NewApiChange(
		RequestParameterDeprecatedSunsetMissingId,
		[]any{param.In, param.Name},
		"",
	)
}
