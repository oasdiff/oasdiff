package commonparams

import (
	"slices"

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
	result := slices.Clone(opParams)
	for _, pathParam := range pathParams {
		if opParams.GetByInAndName(pathParam.Value.In, pathParam.Value.Name) == nil {
			result = append(result, &openapi3.ParameterRef{
				Value: pathParam.Value,
			})
		}
	}
	return result
}
