package checker

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterRemovedId                = "request-parameter-removed" // this is actually the "without deprecation" case but we leave it as is for backward compatibility
	RequestParameterRemovedWithDeprecationId = "request-parameter-removed-with-deprecation"
	RequestParameterSunsetParseId            = "request-parameter-sunset-parse"
	ParameterRemovedBeforeSunsetId           = "request-parameter-removed-before-sunset"
)

func RequestParameterRemovedCheck(diffReport *diff.Diff, operationsSources *diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	if diffReport.PathsDiff == nil {
		return result
	}
	for path, pathItem := range diffReport.PathsDiff.Modified {
		if pathItem.OperationsDiff == nil {
			continue
		}
		for operation, operationItem := range pathItem.OperationsDiff.Modified {
			if operationItem.ParametersDiff == nil {
				continue
			}

			baseSource, revisionSource := operationSources(operationsSources, operationItem.Base, operationItem.Revision)

			opInfo := newOpInfo(
				config,
				operationItem.Revision,
				operationsSources,
				operation,
				path,
			)

			for paramLocation, paramItems := range operationItem.ParametersDiff.Deleted {
				for _, paramName := range paramItems {
					if change, ok := checkParameterRemoval(opInfo, operationItem.Base.Parameters.GetByInAndName(paramLocation, paramName)); ok {
						result = append(result, change.WithSources(baseSource, revisionSource))
					}
				}
			}
		}
	}
	return result
}

func checkParameterRemoval(opInfo opInfo, param *openapi3.Parameter) (ApiChange, bool) {

	if !param.Deprecated {
		return NewApiChange(
			RequestParameterRemovedId,
			opInfo.config,
			[]any{param.In, param.Name},
			commentId(RequestParameterRemovedId),
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		), true
	}

	sunset, ok := getSunset(param.Extensions)
	if !ok {
		return NewApiChange(
			RequestParameterRemovedWithDeprecationId,
			opInfo.config,
			[]any{param.In, param.Name},
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		), true
	}

	date, err := getSunsetDate(sunset)
	if err != nil {
		return getRequestParameterSunsetParse(opInfo, param, err), true
	}

	if civil.DateOf(time.Now()).Before(date) {
		return NewApiChange(
			ParameterRemovedBeforeSunsetId,
			opInfo.config,
			[]any{param.In, param.Name, date},
			"",
			opInfo.operationsSources,
			opInfo.operation,
			opInfo.method,
			opInfo.path,
		), true
	}
	return ApiChange{}, false
}

func getRequestParameterSunsetParse(opInfo opInfo, param *openapi3.Parameter, err error) ApiChange {
	return NewApiChange(
		RequestParameterSunsetParseId,
		opInfo.config,
		[]any{param.In, param.Name, err},
		"",
		opInfo.operationsSources,
		opInfo.operation,
		opInfo.method,
		opInfo.path,
	)
}
