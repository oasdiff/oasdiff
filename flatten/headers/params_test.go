package headers_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/headers"
	"github.com/stretchr/testify/require"
)

func TestLowercaseInPath(t *testing.T) {
	spec := &openapi3.T{}
	spec.Paths = openapi3.NewPathsWithCapacity(1)
	spec.Paths.Set("/path", &openapi3.PathItem{
		Parameters: openapi3.Parameters{
			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name: "X-Header",
					In:   "query",
				},
			},
		},
	})

	headers.Lowercase(spec)

	require.Equal(t, "X-Header", spec.Paths.Find("/path").Parameters[0].Value.Name)
}

func TestLowercaseNonHeader(t *testing.T) {
	spec := &openapi3.T{}
	spec.Paths = openapi3.NewPathsWithCapacity(1)
	spec.Paths.Set("/path", &openapi3.PathItem{
		Parameters: openapi3.Parameters{
			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name: "X-Header",
					In:   "header",
				},
			},
		},
	})

	headers.Lowercase(spec)

	require.Equal(t, "x-header", spec.Paths.Find("/path").Parameters[0].Value.Name)
}

func TestLowercaseInOperation(t *testing.T) {
	spec := &openapi3.T{}
	spec.Paths = openapi3.NewPathsWithCapacity(1)
	spec.Paths.Set("/path", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name: "X-Header",
						In:   "header",
					},
				},
			},
		},
	})

	headers.Lowercase(spec)

	require.Equal(t, "x-header", spec.Paths.Find("/path").GetOperation("GET").Parameters[0].Value.Name)
}

func TestLowercaseInResponse(t *testing.T) {
	spec := &openapi3.T{}
	spec.Paths = openapi3.NewPathsWithCapacity(1)
	spec.Paths.Set("/path", &openapi3.PathItem{
		Get: &openapi3.Operation{},
	})
	spec.Paths.Find("/path").GetOperation("GET").Responses = openapi3.NewResponsesWithCapacity(1)
	spec.Paths.Find("/path").GetOperation("GET").Responses.Set("default", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Headers: openapi3.Headers{
				"X-Header": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{},
					},
				},
			},
		},
	})

	headers.Lowercase(spec)

	require.Contains(t, spec.Paths.Find("/path").GetOperation("GET").Responses.Default().Value.Headers, "x-header")
}
