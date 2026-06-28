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
	requireApiChange(t, checker.ApiChange{
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
	requireApiChanges(t, []checker.ApiChange{
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
	requireApiChange(t, checker.ApiChange{
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
	requireApiChange(t, checker.ApiChange{
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
	requireApiChange(t, checker.ApiChange{
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
//
// Runs the full check set so it also asserts the mechanical artifacts
// (one-of-added, type-generalized) are suppressed. The fixture moves both
// required (foo, bar) and optional (baz) properties into the wrapping, so the
// moved-property suppression is exercised for both.
func TestRequestPropertyOneOfWrappingIsBreaking(t *testing.T) {
	s1, err := open("../data/checker/request_property_one_of_wrapped_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_one_of_wrapped_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.RequestPropertyRemovedId),
		"properties moved into a oneOf wrapping must not be reported as removed (#702)")
	require.False(t, containsId(errs, checker.RequestPropertyBecameOptionalId),
		"oneOf wrapping must not be reported as became-optional (#702)")

	// The mechanical artifacts of the wrapping (the added oneOf alternatives,
	// the top-level type going to "any") are redundant with the single wrapped
	// finding and must be suppressed too.
	require.False(t, containsId(errs, checker.RequestBodyOneOfAddedId),
		"the wrapping's added alternatives must not also be reported as one-of-added (#702)")
	require.False(t, containsId(errs, checker.RequestBodyTypeGeneralizedId),
		"the wrapping's top-level type change to 'any' must not also be reported (#702)")

	// The wrapping must be reported exactly once per request body (not per
	// property), as a breaking error, and nothing else.
	require.Len(t, errs, 1, "a oneOf wrapping must produce exactly one finding (#702)")
	require.Equal(t, checker.RequestBodyWrappedInOneOfId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}
