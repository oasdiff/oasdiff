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
	requireChange(t, errs, checker.NewRequiredRequestPropertyId)

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

// BC: wrapping a concrete request body object into a oneOf of object
// alternatives is a breaking restructuring (#702): under oneOf a previously
// valid payload can match multiple overlapping alternatives and be rejected.
// The moved properties must not be reported as removed, and the wrapping must
// be reported once as request-body-wrapped-in-one-of (ERR), not as
// request-property-became-optional. Reproduces oasdiff/oasdiff#702.
func TestRequestPropertyOneOfWrappingIsBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_property_one_of_wrapped_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_one_of_wrapped_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyUpdatedCheck), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.RequestPropertyRemovedId),
		"properties moved into a oneOf wrapping must not be reported as removed (#702)")
	require.False(t, containsId(errs, checker.RequestPropertyBecameOptionalId),
		"oneOf wrapping must not be reported as became-optional (#702)")

	// The wrapping must be reported exactly once per request body (not per
	// property), as a breaking error.
	wrapped := 0
	for _, e := range errs {
		if e.GetId() == checker.RequestBodyWrappedInOneOfId {
			wrapped++
			require.Equal(t, checker.ERR, e.GetLevel())
		}
	}
	require.Equal(t, 1, wrapped, "the wrapping must be reported once as a breaking error (#702)")
}
