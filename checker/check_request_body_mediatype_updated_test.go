package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a new media type to request body
func TestRequestBodyMediaTypeAdded(t *testing.T) {
	s1, err := open("../data/checker/request_body_media_type_updated_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_media_type_updated_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestBodyMediaTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMediaTypeAddedId,
		Args:        []any{"application/json"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_body_media_type_updated_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// CL: removing media type from request body
func TestRequestBodyMediaTypeRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_body_media_type_updated_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_media_type_updated_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestBodyMediaTypeChangedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyMediaTypeRemovedId,
		Args:        []any{"application/json"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_body_media_type_updated_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
