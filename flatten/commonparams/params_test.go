package commonparams_test

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/commonparams"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
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
	spec.Paths.Find("/path").SetOperation(http.MethodGet, &openapi3.Operation{})

	commonparams.Move(spec)

	require.Empty(t, spec.Paths.Find("/path").Parameters.GetByInAndName("query", "X-Header"))
	require.NotEmpty(t, spec.Paths.Find("/path").Get.Parameters.GetByInAndName("query", "X-Header"))
}
