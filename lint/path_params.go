package lint

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/utils"
)

func PathParamsCheck(source string, s *load.SpecInfo) []*Error {
	result := make([]*Error, 0)

	if s == nil || s.Spec == nil {
		return result
	}

	for path, pathItem := range s.Spec.Paths.Map() {
		pathParamsFromURL := getPathParamsFromURL(path)

		pathParams := utils.StringSet{}
		for _, parameter := range pathItem.Parameters {
			if parameter.Value.In != openapi3.ParameterInPath {
				continue
			}

			if !parameter.Value.Required {
				result = append(result, &Error{
					Id:     "path-param-not-required",
					Level:  LEVEL_ERROR,
					Text:   fmt.Sprintf("path parameter %q should have required=true: %s", parameter.Value.Name, path),
					Source: source,
				})
			}

			pathParams.Add(parameter.Value.Name)
		}

		for method, op := range pathItem.Operations() {
			result = append(result, checkOperationPathParams(pathParamsFromURL, pathParams, path, method, op, source)...)
		}
	}

	return result
}

func getPathParamsFromURL(path string) utils.StringSet {
	_, _, pathParams := utils.NormalizeTemplatedPath(path)
	return utils.StringSetFromSlice(pathParams)
}

func checkOperationPathParams(pathParamsFromURL, pathParams utils.StringSet, path, method string, op *openapi3.Operation, source string) []*Error {
	result := make([]*Error, 0)

	opParams := utils.StringSet{}
	for _, parameter := range op.Parameters {
		if parameter.Value.In != openapi3.ParameterInPath {
			continue
		}

		if !parameter.Value.Required {
			result = append(result, &Error{
				Id:     "path-param-not-required",
				Level:  LEVEL_ERROR,
				Text:   fmt.Sprintf("path parameter %q should have required=true: %s %s", parameter.Value.Name, method, path),
				Source: source,
			})
		}

		opParams.Add(parameter.Value.Name)
	}

	for param := range pathParams.Plus(opParams).Minus(pathParamsFromURL) {
		result = append(result, &Error{
			Id:     "path-param-extra",
			Level:  LEVEL_ERROR,
			Text:   getParamMissingText(opParams, param, method, path),
			Source: source,
		})
	}

	for param := range pathParamsFromURL.Minus(pathParams).Minus(opParams) {
		result = append(result, &Error{
			Id:     "path-param-missing",
			Level:  LEVEL_WARN,
			Text:   fmt.Sprintf("path parameter %q appears in the URL path but is missing from the parameters section of the path and operation: %s %s", param, method, path),
			Source: source,
		})
	}

	for param := range pathParams.Intersection(opParams) {
		result = append(result, &Error{
			Id:     "path-param-duplicate",
			Level:  LEVEL_WARN,
			Text:   fmt.Sprintf("path parameter %q is defined both in path and in operation: %s %s", param, method, path),
			Source: source,
		})
	}

	return result
}

func getParamMissingText(opParams utils.StringSet, param, method, path string) string {
	if opParams.Contains(param) {
		return fmt.Sprintf("path parameter %q appears in the parameters section of the operation but is missing in the URL: %s %s", param, method, path)
	}
	return fmt.Sprintf("path parameter %q appears in the parameters section of the path but is missing in the URL: %s %s", param, method, path)
}
