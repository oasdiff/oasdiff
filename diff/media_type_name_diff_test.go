package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestMediaTypeNameDiff(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/add_new_media_type_revision.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/checker/add_new_media_type_name_modified.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	nd := d.PathsDiff.Modified["/api/v1.0/groups"].OperationsDiff.Modified["POST"].ResponsesDiff.Modified["200"].ContentDiff.MediaTypeModified["application/problem+json"].NameDiff

	require.Equal(t, "application/json", nd.NameDiff.From)
	require.Equal(t, "application/problem+json", nd.NameDiff.To)
	require.Nil(t, nd.ParametersDiff)
	require.Nil(t, nd.TypeDiff)
	require.Equal(t, "json", nd.SubtypeDiff.From)
	require.Equal(t, "problem", nd.SubtypeDiff.To)
	require.Equal(t, "", nd.SuffixDiff.From)
	require.Equal(t, "json", nd.SuffixDiff.To)
	require.Nil(t, nd.ParametersDiff)
	require.True(t, nd.IsContained())
}
