package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// adding a required property to response body is detected
func TestResponseRequiredPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseRequiredPropertyAddedId,
		Args:        []any{"data/new", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// removing an existent property that was required in response body is detected
func TestResponseRequiredPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Required = []string{"name", "id"}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseRequiredPropertyRemovedId,
		Args:        []any{"data/new", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// adding a required write-only property to response body is detected
func TestResponseRequiredWriteOnlyPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["new"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{

		Id:          checker.ResponseRequiredWriteOnlyPropertyAddedId,
		Args:        []any{"data/new", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// removing a required write-only property that was required in response body is detected
func TestResponseRequiredWriteOnlyPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)

	s1.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["new"].Value.WriteOnly = true
	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Required = []string{"name", "id"}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseRequiredWriteOnlyPropertyRemovedId,
		Args:        []any{"data/new", "200"},
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs)
}

// wrapping a concrete object response body into a oneOf (#702) is a
// breaking restructuring: a field that was previously guaranteed in the
// response may now be absent depending on which alternative matches. The moved
// properties must not be reported as removed, and the wrapping must be reported
// once as response-body-wrapped-in-one-of (ERR).
//
// The fixture moves both required (foo, bar) and optional (baz) properties into
// the wrapping, so this exercises the moved-property suppression in both
// ResponseRequiredPropertyUpdatedCheck and ResponseOptionalPropertyUpdatedCheck.
func TestResponsePropertyOneOfWrappingIsBreaking(t *testing.T) {
	s1, err := open("../data/checker/response_property_one_of_wrapped_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_wrapped_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.False(t, containsId(errs, checker.ResponseRequiredPropertyRemovedId),
		"properties moved into a oneOf wrapping must not be reported as removed (#702)")
	require.False(t, containsId(errs, checker.ResponseOptionalPropertyRemovedId),
		"properties moved into a oneOf wrapping must not be reported as removed (#702)")

	// The mechanical artifacts of the wrapping (the added oneOf alternatives,
	// the top-level type going to "any") are redundant with the single wrapped
	// finding and must be suppressed too.
	require.False(t, containsId(errs, checker.ResponseBodyOneOfAddedId),
		"the wrapping's added alternatives must not also be reported as one-of-added (#702)")
	require.False(t, containsId(errs, checker.ResponseBodyTypeChangedId),
		"the wrapping's top-level type change to 'any' must not also be reported (#702)")

	// The wrapping must be reported exactly once per response body (not per
	// property), as a breaking error, and nothing else.
	require.Len(t, errs, 1, "a oneOf wrapping must produce exactly one finding (#702)")
	require.Equal(t, checker.ResponseBodyWrappedInOneOfId, errs[0].GetId())
	require.Equal(t, checker.ERR, errs[0].GetLevel())
}
