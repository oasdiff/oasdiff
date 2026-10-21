package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func loadCommonParams(t *testing.T, file string) *load.SpecInfo {
	t.Helper()
	s, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/common-params/"+file))
	require.NoError(t, err)
	return s
}

// moving parameters from the path item to the operation is a change to the
// document but not to the parameters the operation has
func TestParametersContext_Moved(t *testing.T) {
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), loadCommonParams(t, "params_in_path.yaml"), loadCommonParams(t, "params_in_op.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, d.PathsDiff.Modified["/admin/v0/abc/{id}"].ParametersDiff.Deleted)
	require.Empty(t, d.ParametersContext.Operations)
}

func TestParametersContext_Deleted(t *testing.T) {
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), loadCommonParams(t, "params_in_path.yaml"), loadCommonParams(t, "no_params.yaml"))
	require.NoError(t, err)
	require.Len(t, d.ParametersContext.Operations, 1)
	operation := d.ParametersContext.Operations[0]
	require.Equal(t, "/admin/v0/abc/{id}", operation.Path)
	require.Equal(t, "GET", operation.Method)
	require.Equal(t, diff.ParamNamesByLocation{"header": {"tenant-id"}, "path": {"id"}}, operation.ParametersDiff.Deleted)
	require.Len(t, operation.BaseParameters, 2)
	require.Empty(t, operation.RevisionParameters)
}

func TestParametersContext_Composed(t *testing.T) {
	d, _, err := diff.GetPathsDiff(diff.NewConfig(), []*load.SpecInfo{loadCommonParams(t, "params_in_path.yaml")}, []*load.SpecInfo{loadCommonParams(t, "no_params.yaml")})
	require.NoError(t, err)
	require.Len(t, d.ParametersContext.Operations, 1)
	require.Equal(t, diff.ParamNamesByLocation{"header": {"tenant-id"}, "path": {"id"}}, d.ParametersContext.Operations[0].ParametersDiff.Deleted)
}

func TestParametersContext_Webhook(t *testing.T) {
	spec := func(required bool) *load.SpecInfo {
		s, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/checker/parameter_positions.yaml"))
		require.NoError(t, err)
		s.Spec.Webhooks["newPet"].Parameters = openapi3.Parameters{{Value: &openapi3.Parameter{Name: "limit", In: openapi3.ParameterInQuery, Required: required, Schema: openapi3.NewIntegerSchema().NewRef()}}}
		return s
	}
	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), spec(false), spec(true))
	require.NoError(t, err)
	require.Empty(t, d.ParametersContext.Operations)
	require.Len(t, d.ParametersContext.Webhooks, 1)
	require.Equal(t, "newPet", d.ParametersContext.Webhooks[0].Path)
	require.Equal(t, "POST", d.ParametersContext.Webhooks[0].Method)
	require.NotNil(t, d.ParametersContext.Webhooks[0].ParametersDiff.Modified["query"]["limit"].RequiredDiff)
}
