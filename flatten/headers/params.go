package headers

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Lowercase replaces header names to lowercase
func Lowercase(spec *openapi3.T) {
	lowerHeaderNames(spec)
}

func lowerHeaderNames(spec *openapi3.T) {
	for _, path := range spec.Paths.Map() {

		for _, paramRef := range path.Parameters {
			lowerHeaderName(paramRef.Value)
		}

		for _, op := range path.Operations() {
			for _, paramRef := range op.Parameters {
				lowerHeaderName(paramRef.Value)
			}

			for _, responseRef := range op.Responses.Map() {
				responseRef.Value.Headers = lowerResponseHeaders(responseRef.Value.Headers)
			}
		}
	}
}

func lowerHeaderName(param *openapi3.Parameter) {
	if param.In != openapi3.ParameterInHeader {
		return
	}

	param.Name = strings.ToLower(param.Name)
}

// lowerResponseHeaders returns a new headers map with lowercase keys
// Explanation:
// According to the openapi spec, the Header Object follows the structure of the Parameter Object with the following changes:
// 1. name MUST NOT be specified, it is given in the corresponding headers map.
// 2. in MUST NOT be specified, it is implicitly in header.
// See: https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#header-object
func lowerResponseHeaders(headers openapi3.Headers) openapi3.Headers {
	result := make(openapi3.Headers, len(headers))
	for k, v := range headers {
		result[strings.ToLower(k)] = v
	}
	return result
}
