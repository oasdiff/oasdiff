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

func TestMerge(t *testing.T) {
	pathLimit := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "limit", In: "query"}}
	pathTenant := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "tenant", In: "header"}}
	opLimit := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "limit", In: "query", Required: true}}
	opParams := openapi3.Parameters{opLimit}

	merged := commonparams.Merge(openapi3.Parameters{pathLimit, pathTenant}, opParams)

	// the operation's own limit overrides the path's
	require.Same(t, opLimit.Value, merged.GetByInAndName("query", "limit"))
	require.Same(t, pathTenant.Value, merged.GetByInAndName("header", "tenant"))
	require.Len(t, merged, 2)
	require.Equal(t, openapi3.Parameters{opLimit}, opParams)
}

func TestMergeIgnoringHeaderCase(t *testing.T) {
	pathTrace := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "X-Trace", In: "header"}}
	pathQuery := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "Limit", In: "query"}}
	opTrace := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "x-trace", In: "header"}}
	opQuery := &openapi3.ParameterRef{Value: &openapi3.Parameter{Name: "limit", In: "query"}}
	pathParams := openapi3.Parameters{pathTrace, pathQuery}
	opParams := openapi3.Parameters{opTrace, opQuery}

	// the operation's header overrides the path's, whose name differs in case;
	// query parameter names are case-sensitive, so both are kept
	merged := commonparams.MergeIgnoringHeaderCase(pathParams, opParams)
	require.Len(t, merged, 3)
	require.Nil(t, merged.GetByInAndName("header", "X-Trace"))
	require.Same(t, pathQuery.Value, merged.GetByInAndName("query", "Limit"))

	// Merge compares names as written
	require.Len(t, commonparams.Merge(pathParams, opParams), 4)
}
