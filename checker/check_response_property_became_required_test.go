package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing optional response property to required
func TestResponsePropertyBecameRequiredlCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_became_optional_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_became_optional_base.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameRequiredCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameRequiredId,
		Args:        []any{"data/name", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_became_optional_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// changing optional response property to required with source tracking
func TestResponsePropertyBecameRequired_WithSources(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/response_property_became_optional_revision.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_became_optional_base.yaml", loader)
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameRequiredCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.ResponsePropertyBecameRequiredId)

	// Base has no 'required' list in YAML, so baseSource is nil
	require.Empty(t, errs[0].GetBaseSource())
	// Revision has 'required' list, so revisionSource points to it (sequence field)
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/response_property_became_optional_base.yaml", errs[0].GetRevisionSource().File)
}

// changing optional response write-only property to required
func TestResponseWriteOnlyPropertyBecameRequiredCheck(t *testing.T) {
	s1, err := open("../data/checker/response_property_became_optional_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_became_optional_base.yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["name"].Value.WriteOnly = true

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyBecameRequiredCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseWriteOnlyPropertyBecameRequiredId,
		Args:        []any{"data/name", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_property_became_optional_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}
