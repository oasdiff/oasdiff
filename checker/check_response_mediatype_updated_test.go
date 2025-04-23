package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a new media type to response
func TestAddNewMediaType(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseMediaTypeAddedId,
		Args:        []any{"application/xml", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/add_new_media_type_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: removing a media type from response
func TestDeleteNewMediaType(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseMediaTypeRemovedId,
		Args:        []any{"application/xml", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/add_new_media_type_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}
