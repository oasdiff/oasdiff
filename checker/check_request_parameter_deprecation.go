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
	if diffReport.PathsDiff == nil {
		return result
	}

	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationDiff := range pathItem.OperationsDiff.Modified {
			if operationDiff.ParametersDiff == nil {
				continue
			}

			op := pathItem.Revision.GetOperation(operation)
			opInfo := newOpInfo(config, op, operationsSources, operation, path)

			for paramLocation, paramItems := range operationDiff.ParametersDiff.Modified {
				for paramName, paramItem := range paramItems {

					if paramItem.DeprecatedDiff == nil {
						continue
					}

					param := paramItem.Revision

					if paramItem.DeprecatedDiff.To == nil || paramItem.DeprecatedDiff.To == false {
						// not breaking changes
						stability, _ := getStabilityLevel(op.Extensions)
						details := formatDeprecationDetails(nil, stability)
						result = append(result, NewApiChange(
							RequestParameterReactivatedId,
							config,
							[]any{paramLocation, paramName},
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

					sunset, ok := getSunset(param.Extensions)
					if !ok {
						// if deprecation policy is defined and sunset is missing, it's a breaking change
						if deprecationDays > 0 {
							result = append(result, getParameterDeprecatedSunsetMissing(opInfo, param))
						} else {
							// no policy, report deprecation without sunset as INFO
							details := formatDeprecationDetails(nil, stability)
							result = append(result, NewApiChange(
								RequestParameterDeprecatedId,
								config,
								[]any{paramLocation, paramName},
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
							RequestParameterSunsetParseId,
							config,
							[]any{param.In, param.Name, err},
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
							RequestParameterSunsetDateTooSmallId,
							config,
							[]any{param.In, param.Name, date, deprecationDays},
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
						RequestParameterDeprecatedId,
						config,
						[]any{paramLocation, paramName},
						details,
						operationsSources,
						op,
						operation,
						path,
					))
				}
			}
		}
	}

	return result
}

func getParameterDeprecatedSunsetMissing(opInfo opInfo, param *openapi3.Parameter) Change {
	return NewApiChange(
		RequestParameterDeprecatedSunsetMissingId,
		opInfo.config,
		[]any{param.In, param.Name},
		"",
		opInfo.operationsSources,
		opInfo.operation,
		opInfo.method,
		opInfo.path,
	)
}
