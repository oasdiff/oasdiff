package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: removing an existing operation id
func TestOperationIdRemoved(t *testing.T) {
	s1, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.OperationID = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIOperationIdUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.APIOperationIdRemovedId,
		Args:        []any{"createOneGroup", ""},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/operation_id_removed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: updating an existing operation id
func TestOperationIdUpdated(t *testing.T) {
	s1, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.OperationID = "newOperationId"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIOperationIdUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.APIOperationIdRemovedId,
		Args:        []any{"createOneGroup", "newOperationId"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/operation_id_removed_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])

	require.Equal(t, "api operation id 'createOneGroup' removed and replaced with 'newOperationId'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// CL: removing an existing operation id with source tracking
func TestOperationIdRemoved_WithSources(t *testing.T) {
	enableOriginTracking(t)
	s1, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/operation_id_removed_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.OperationID = ""

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIOperationIdUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIOperationIdRemovedId, errs[0].GetId())

	// Verify source tracking
	require.NotEmpty(t, errs[0].GetBaseSource())
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/operation_id_removed_base.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, "../data/checker/operation_id_removed_base.yaml", errs[0].GetRevisionSource().File)
	require.NotZero(t, errs[0].GetBaseSource().Line)
	require.NotZero(t, errs[0].GetRevisionSource().Line)
}

// CL: adding a new operation id
func TestOperationIdAdded(t *testing.T) {
	s1, err := open("../data/checker/operation_id_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/operation_id_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.OperationID = "NewOperationId"

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIOperationIdUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.APIOperationIdAddId,
		Args:        []any{"NewOperationId"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/operation_id_added_base.yaml"),
		OperationId: "NewOperationId",
	}, errs[0])

	require.Equal(t, "api operation id 'NewOperationId' was added", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
