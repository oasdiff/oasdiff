package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a new required request property
func TestRequiredRequestPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.NewRequiredRequestPropertyId,
		Args:        []any{"description"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_added_revision.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: adding two new request properties, one required, one optional
func TestRequiredRequestPropertiesAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_revision2.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.NewRequiredRequestPropertyId,
			Args:        []any{"description"},
			Level:       checker.ERR,
			Operation:   "POST",
			Path:        "/products",
			Source:      load.NewSource("../data/checker/request_property_added_revision2.yaml"),
			OperationId: "addProduct",
		},
		{
			Id:          checker.NewOptionalRequestPropertyId,
			Args:        []any{"info"},
			Level:       checker.INFO,
			Operation:   "POST",
			Path:        "/products",
			Source:      load.NewSource("../data/checker/request_property_added_revision2.yaml"),
			OperationId: "addProduct",
		}}, errs)
}

// CL: adding a new required request property with source tracking
func TestRequiredRequestPropertyAdded_WithSources(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/request_property_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.NewRequiredRequestPropertyId, errs[0].GetId())

	// Added property: base source is nil (property doesn't exist in base), revision source points to the property
	require.Empty(t, errs[0].GetBaseSource())
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/request_property_added_revision.yaml", errs[0].GetRevisionSource().File)
	require.NotZero(t, errs[0].GetRevisionSource().Line)
}

// CL: adding a new optional request property
func TestRequiredOptionalPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Paths.Value("/products").Post.RequestBody.Value.Content["application/json"].Schema.Value.Required = []string{"name"}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.NewOptionalRequestPropertyId,
		Args:        []any{"description"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_added_revision.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: removing a required request property
func TestRequiredRequestPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.RequestPropertyRemovedId,
		Args:        []any{"description"},
		Level:       checker.WARN,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_added_base.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// CL: adding a new required request property with a default value
func TestRequiredRequestPropertyAddedWithDefault(t *testing.T) {
	s1, err := open("../data/checker/request_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_added_with_default.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.NewRequiredRequestPropertyWithDefaultId,
		Args:        []any{"description"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/products",
		Source:      load.NewSource("../data/checker/request_property_added_with_default.yaml"),
		OperationId: "addProduct",
	}, errs[0])
}

// BC: wrapping a request body object into a oneOf of object alternatives is not
// a removal of its properties: they move into the alternatives. The spurious
// request-property-removed findings must be suppressed, and a base-required
// property that is no longer required in every alternative is reported as
// became-optional. Reproduces oasdiff/oasdiff#702.
func TestRequestPropertyOneOfWrappingNotRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_one_of_wrapped_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_one_of_wrapped_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.RequestPropertyRemovedId),
		"properties moved into a oneOf wrapping must not be reported as removed (#702)")

	var becameOptional []string
	for _, e := range errs {
		require.Equal(t, checker.RequestPropertyBecameOptionalId, e.GetId(),
			"the only findings from wrapping should be became-optional")
		require.Equal(t, checker.INFO, e.GetLevel())
		becameOptional = append(becameOptional, e.GetArgs()[0].(string))
	}
	require.ElementsMatch(t, []string{"foo", "bar"}, becameOptional)
}
