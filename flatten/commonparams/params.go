package commonparams

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Move moves common parameters to the operations under the path
func Move(spec *openapi3.T) {
	for _, path := range spec.Paths.Map() {
		for _, op := range path.Operations() {
			op.Parameters = Merge(path.Parameters, op.Parameters)
		}
		path.Parameters = nil
	}
}

// Merge returns the parameters an operation has: its own, followed by the
// path-level ones it doesn't override by location and name. It doesn't modify
// either list.
func Merge(pathParams, opParams openapi3.Parameters) openapi3.Parameters {
	return merge(pathParams, opParams, func(pathParam *openapi3.Parameter) bool {
		return opParams.GetByInAndName(pathParam.In, pathParam.Name) != nil
	})
}

// MergeIgnoringHeaderCase is Merge, except that an operation's header
// parameter overrides a path-level one whose name differs only in case, since
// HTTP header names are case-insensitive.
func MergeIgnoringHeaderCase(pathParams, opParams openapi3.Parameters) openapi3.Parameters {
	return merge(pathParams, opParams, func(pathParam *openapi3.Parameter) bool {
		if pathParam.In != openapi3.ParameterInHeader {
			return opParams.GetByInAndName(pathParam.In, pathParam.Name) != nil
		}
		return slices.ContainsFunc(opParams, func(opParam *openapi3.ParameterRef) bool {
			return opParam.Value != nil && opParam.Value.In == openapi3.ParameterInHeader && strings.EqualFold(opParam.Value.Name, pathParam.Name)
		})
	})
}

func merge(pathParams, opParams openapi3.Parameters, overridden func(pathParam *openapi3.Parameter) bool) openapi3.Parameters {
	result := slices.Clone(opParams)
	for _, pathParam := range pathParams {
		if !overridden(pathParam.Value) {
			result = append(result, &openapi3.ParameterRef{
				Value: pathParam.Value,
			})
		}
	}
	return result
}
