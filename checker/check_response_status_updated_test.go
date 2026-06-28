package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a success response status
func TestResponseSuccessStatusAdded(t *testing.T) {
	s1, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)

	// Add new success response
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Set("201", s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("200"))

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseSuccessStatusUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.ResponseSuccessStatusAddedId,
		Args:        []any{"201"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_status_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: adding a non-success response status
func TestResponseNonSuccessStatusAdded(t *testing.T) {
	s1, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)

	// Add new non-success response
	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Set("400", s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Value("409"))

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseNonSuccessStatusUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.ResponseNonSuccessStatusAddedId,
		Args:        []any{"400"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_status_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: removing a non-success response status
func TestResponseNonSuccessStatusRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Delete("409")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseNonSuccessStatusUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.ResponseNonSuccessStatusRemovedId,
		Args:        []any{"409"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_status_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// BC: removing a success status is breaking
func TestResponseSuccessStatusRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_status_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.Responses.Delete("200")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseSuccessStatusUpdatedCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	requireApiChange(t, checker.ApiChange{
		Id:          checker.ResponseSuccessStatusRemovedId,
		Args:        []any{"200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_status_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
