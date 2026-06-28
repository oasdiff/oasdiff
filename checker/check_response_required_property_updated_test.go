package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding a required property to response body is detected
func TestResponseRequiredPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          checker.ResponseRequiredPropertyAddedId,
		Args:        []any{"data/new", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: removing an existent property that was required in response body is detected
func TestResponseRequiredPropertyRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Required = []string{"name", "id"}
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseRequiredPropertyRemovedId,
		Args:        []any{"data/new", "200"},
		Level:       checker.ERR,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: adding a required write-only property to response body is detected
func TestResponseRequiredWriteOnlyPropertyAdded(t *testing.T) {
	s1, err := open("../data/checker/response_required_property_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_required_property_added_revision.yaml")
	require.NoError(t, err)

	s2.Spec.Components.Schemas["GroupView"].Value.Properties["data"].Value.Properties["new"].Value.WriteOnly = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseRequiredPropertyUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{

		Id:          checker.ResponseRequiredWriteOnlyPropertyAddedId,
		Args:        []any{"data/new", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_revision.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: removing a required write-only property that was required in response body is detected
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
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:          checker.ResponseRequiredWriteOnlyPropertyRemovedId,
		Args:        []any{"data/new", "200"},
		Level:       checker.INFO,
		Operation:   "POST",
		Path:        "/api/v1.0/groups",
		Source:      load.NewSource("../data/checker/response_required_property_added_base.yaml"),
		OperationId: "createOneGroup",
	}, errs[0])
}

// CL: wrapping a concrete object response body into a oneOf (#702) is a
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

	// The wrapping must be reported exactly once per response body (not per
	// property), as a breaking error.
	wrapped := 0
	for _, e := range errs {
		if e.GetId() == checker.ResponseBodyWrappedInOneOfId {
			wrapped++
			require.Equal(t, checker.ERR, e.GetLevel())
		}
	}
	require.Equal(t, 1, wrapped, "the wrapping must be reported once as a breaking error (#702)")
}
