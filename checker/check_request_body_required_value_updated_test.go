package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing request's body to required is breaking
func TestRequestBodyBecameRequired(t *testing.T) {
	s1, err := open("../data/checker/request_body_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_became_required_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestBodyRequiredUpdatedCheck), d, osm)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyBecameRequiredId,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_body_became_required_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing request's body to optional
func TestRequestBodyBecameOptional(t *testing.T) {
	s1, err := open("../data/checker/request_body_became_optional_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_became_optional_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/v1.0/groups").Post.RequestBody.Value.Required = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestBodyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyBecameOptionalId,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/request_body_became_optional_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
