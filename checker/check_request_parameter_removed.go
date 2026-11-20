package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/checker/internal/location"
	"github.com/oasdiff/oasdiff/diff"
)

const (
	RequestParameterRemovedId                = "request-parameter-removed" // this is actually the "without deprecation" case but we leave it as is for backward compatibility
	RequestParameterRemovedWithDeprecationId = "request-parameter-removed-with-deprecation"
	RequestParameterSunsetParseId            = "request-parameter-sunset-parse"
	ParameterRemovedBeforeSunsetId           = "request-parameter-removed-before-sunset"
)

func RequestParameterRemovedCheck(diffReport *diff.Diff, operationsSources diff.OperationsSourcesMap, config *Config) Changes {
	result := make(Changes, 0)
	walkOperationsParameters(diffReport, operationsSources, config, func(o operationParams) {
		for paramLocation, paramItems := range o.ParametersDiff.Deleted {
			for _, paramName := range paramItems {
				param := o.BaseParameters.GetByInAndName(paramLocation, paramName)
				if change, ok := checkParameterRemoval(o.opInfo, param); ok {
					baseSource := location.ParameterSource(operationsSources, o.Base, param)
					result = append(result, change.WithSources(baseSource, nil))
				}
			}
		}
	})
	return result
}

func checkParameterRemoval(opInfo opInfo, param *openapi3.Parameter) (ApiChange, bool) {
	id, extra, ok := deprecation.Removal(param.Deprecated, param.Extensions, deprecation.RemovalIds{
		Removed:         RequestParameterRemovedId,
		WithDeprecation: RequestParameterRemovedWithDeprecationId,
		BeforeSunset:    ParameterRemovedBeforeSunsetId,
		SunsetParse:     RequestParameterSunsetParseId,
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
