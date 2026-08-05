package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// OpenAPI 3.2 adds the QUERY path item field and additionalOperations, a map of
// custom http methods. Both hold operations like any other, so the diff
// compares them alongside the methods with a dedicated field.

func oas32Diff(t *testing.T) *diff.Diff {
	t.Helper()

	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/oas32/base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/oas32/revision.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return d
}

func TestDiff_QueryOperationModified(t *testing.T) {
	operationsDiff := oas32Diff(t).PathsDiff.Modified["/search"].OperationsDiff

	require.Contains(t, operationsDiff.Modified, openapi3.MethodQuery)
	require.NotEmpty(t, operationsDiff.Modified[openapi3.MethodQuery].ResponsesDiff.Deleted)
	require.NotEmpty(t, operationsDiff.Modified[openapi3.MethodQuery].RequestBodyDiff)
}

func TestDiff_CustomMethodAddedDeletedAndModified(t *testing.T) {
	operationsDiff := oas32Diff(t).PathsDiff.Modified["/search"].OperationsDiff

	require.Contains(t, operationsDiff.Added, "MOVE")
	require.Contains(t, operationsDiff.Deleted, "PURGE")
	require.Contains(t, operationsDiff.Modified, "COPY")
	require.NotNil(t, operationsDiff.Modified["COPY"].DeprecatedDiff)
}

// Custom methods are reported after the fixed-field methods and sorted among
// themselves, so repeated runs produce the same output despite the map.
func TestDiff_CustomMethodOrderIsStable(t *testing.T) {
	for range 10 {
		operationsDiff := oas32Diff(t).PathsDiff.Modified["/search"].OperationsDiff
		require.Equal(t, []string{"MOVE"}, operationsDiff.Added)
		require.Equal(t, []string{"PURGE"}, operationsDiff.Deleted)
	}
}

// An added or deleted path carries its QUERY and custom-method operations into
// the endpoints diff, the same as any other method.
func TestDiff_CustomMethodEndpointsOnAddedAndDeletedPaths(t *testing.T) {
	d := oas32Diff(t)

	require.Contains(t, d.EndpointsDiff.Deleted, diff.Endpoint{Method: openapi3.MethodQuery, Path: "/gone"})
	require.Contains(t, d.EndpointsDiff.Deleted, diff.Endpoint{Method: "LOCK", Path: "/gone"})
	require.Contains(t, d.EndpointsDiff.Added, diff.Endpoint{Method: openapi3.MethodQuery, Path: "/new"})
}

// An additionalOperations key that repeats a fixed field is invalid OpenAPI;
// validate reports it. The diff must not also count the method twice, which
// would report it as added on top of comparing the fixed field.
func TestDiff_CustomMethodShadowingFixedFieldIsNotDoubleCounted(t *testing.T) {
	loader := openapi3.NewLoader()

	s1, err := load.NewSpecInfo(loader, load.NewSource("../data/oas32/shadowed-method-base.yaml"))
	require.NoError(t, err)

	s2, err := load.NewSpecInfo(loader, load.NewSource("../data/oas32/shadowed-method.yaml"))
	require.NoError(t, err)

	d, _, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	require.True(t, d.Empty(), "the shadowed key is not an operation of its own")
}
