package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestDiff_CommonParamsDeleted(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/params_in_path.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/no_params.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, d.EndpointsDiff.Modified[diff.Endpoint{Method: "GET", Path: "/admin/v0/abc/{id}"}].ParametersDiff.Deleted)
}

func TestDiff_CommonParamsMoved(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/params_in_path.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/params_in_op.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, d)
}

func TestDiff_CommonParamsAdded(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/no_params.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/common-params/params_in_path.yaml"), load.WithFlattenParams())
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotEmpty(t, d.EndpointsDiff.Modified[diff.Endpoint{Method: "GET", Path: "/admin/v0/abc/{id}"}].ParametersDiff.Added)
}
