package commonparams

import "github.com/getkin/kin-openapi/openapi3"

// Move moves common parameters to the operations under the path
func Move(spec *openapi3.T) {
	moveParams(spec)
}

func moveParams(spec *openapi3.T) {
	for _, path := range spec.Paths.Map() {
		for _, op := range path.Operations() {
			addParams(op, path.Parameters)
		}
		path.Parameters = nil
	}
}

func addParams(op *openapi3.Operation, pathParams openapi3.Parameters) {
	for _, pathParam := range pathParams {
		op.Parameters = addParam(op.Parameters, pathParam.Value)
	}
}

func addParam(opParams openapi3.Parameters, pathParam *openapi3.Parameter) openapi3.Parameters {
	if opParams.GetByInAndName(pathParam.In, pathParam.Name) == nil {
		opParams = append(opParams, &openapi3.ParameterRef{
			Value: pathParam,
		})
	}
	return opParams
}
