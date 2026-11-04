package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/location"
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

			opInfo := newOpInfo(
				config,
				operationItem.Revision,
				operationsSources,
				operation,
				path,
			)

			for paramLocation, paramItems := range operationItem.ParametersDiff.Deleted {
				for _, paramName := range paramItems {
					param := operationItem.Base.Parameters.GetByInAndName(paramLocation, paramName)
					if change, ok := checkParameterRemoval(opInfo, param); ok {
						baseSource := location.ParameterSource(operationsSources, operationItem.Base, param)
						result = append(result, change.WithSources(baseSource, nil))
					}
				}
			}
		}
	}
	return result
}

func checkParameterRemoval(opInfo opInfo, param *openapi3.Parameter) (ApiChange, bool) {
	id, extra, ok := removalId(param.Deprecated, param.Extensions, removalIds{
		removed:         RequestParameterRemovedId,
		withDeprecation: RequestParameterRemovedWithDeprecationId,
		beforeSunset:    ParameterRemovedBeforeSunsetId,
		sunsetParse:     RequestParameterSunsetParseId,
	})
	if !ok {
		return ApiChange{}, false
	}

	comment := ""
	if id == RequestParameterRemovedId {
		comment = commentId(RequestParameterRemovedId)
	}
	return opInfo.NewApiChange(id, append([]any{param.In, param.Name}, extra...), comment), true
}

func getRequestParameterSunsetParse(opInfo opInfo, param *openapi3.Parameter, err error) ApiChange {
	return opInfo.NewApiChange(
		RequestParameterSunsetParseId,
		[]any{param.In, param.Name, err},
		"",
	)
}
