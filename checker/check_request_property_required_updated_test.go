package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// changing request property required value to true
func TestRequestPropertyMarkedRequired(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameRequiredId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_required_base.yaml"),
		OperationId: "addProduct",
	}, errs)
}

// changing request property required value to false
func TestRequestPropertyMarkedOptional(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameOptionalId,
		Args:        []any{"name"},
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_required_base.yaml"),
		OperationId: "addProduct",
	}, errs)
}

// making request property required, while also giving it a default value
func TestRequestPropertyWithDefaultMarkedRequired(t *testing.T) {
	s1, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_became_required_base.yaml")
	require.NoError(t, err)

	s1.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{""}
	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Properties["name"].Value.Default = "default"
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyRequiredUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestPropertyBecameRequiredWithDefaultId,
		Args:        []any{"name"},
		Comment:     checker.RequiredRequestPropertyWithDefaultCommentId,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_became_required_base.yaml"),
		OperationId: "addProduct",
	}, errs)
}
