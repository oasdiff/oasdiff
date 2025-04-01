package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// BC: removing request body is breaking
func TestRequestBodyRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_body_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestBodyRemovedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestBodyRemovedId,
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/test",
		Source:      load.NewSource("../data/checker/request_body_removed_revision.yaml"),
		OperationId: "testOp",
	}, errs[0])
}
