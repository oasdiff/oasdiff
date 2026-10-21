package checker_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CheckBackwardCompatibilityUntilLevel must not mutate the caller's diff.
// Regression for #904.
func TestBreaking_CheckDoesNotMutateCallerDiff(t *testing.T) {
	s1, err := open("../data/checker/clone_diff_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/clone_diff_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// The stability filter drops the alpha paths and operations, and the
	// webhook is merged into PathsDiff.Modified. A slice truncated in place
	// only changes the caller's copy when a dropped entry precedes a kept
	// one, so the deleted paths are put in that order (the deleted methods
	// already are: DELETE is alpha, PATCH is stable).
	require.NotEmpty(t, d.WebhooksDiff.Modified)
	slices.Sort(d.PathsDiff.Deleted)
	require.Equal(t, []string{"/alpha-gone", "/stable-gone"}, d.PathsDiff.Deleted)
	pets := d.PathsDiff.Modified["/pets"].OperationsDiff
	require.Equal(t, []string{"DELETE", "PATCH"}, pets.Deleted)
	require.NotEmpty(t, pets.Modified)

	pathsDeleted := slices.Clone(d.PathsDiff.Deleted)
	pathsModified := slices.Sorted(maps.Keys(d.PathsDiff.Modified))
	operationsDeleted := slices.Clone(pets.Deleted)
	operationsModified := slices.Sorted(maps.Keys(pets.Modified))

	_ = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Equal(t, pathsDeleted, d.PathsDiff.Deleted)
	require.Equal(t, pathsModified, slices.Sorted(maps.Keys(d.PathsDiff.Modified)))
	require.Same(t, pets, d.PathsDiff.Modified["/pets"].OperationsDiff)
	require.Equal(t, operationsDeleted, pets.Deleted)
	require.Equal(t, operationsModified, slices.Sorted(maps.Keys(pets.Modified)))
}

func TestBreaking_CheckDoesNotMutateCallerSecurityContext(t *testing.T) {
	s1, err := open("../data/checker/api_security_anonymous_alpha_global_api_key.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/api_security_anonymous_alpha_global_anonymous.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d.SecurityContext)
	before := slices.Clone(d.SecurityContext.Operations)

	_ = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Equal(t, before, d.SecurityContext.Operations, "the alpha operation was filtered out of the caller's security context")
}

func TestBreaking_CheckDoesNotMutateCallerParametersContext(t *testing.T) {
	spec := func(required bool) *load.SpecInfo {
		s, err := open("../data/checker/parameter_positions.yaml")
		require.NoError(t, err)
		pathItem := s.Spec.Paths.Value("/pets")
		pathItem.Parameters = openapi3.Parameters{{Value: &openapi3.Parameter{Name: "limit", In: openapi3.ParameterInQuery, Required: required, Schema: openapi3.NewIntegerSchema().NewRef()}}}
		pathItem.Get.Extensions = map[string]any{"x-stability-level": "alpha"}
		return s
	}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), spec(false), spec(true))
	require.NoError(t, err)
	require.Len(t, d.ParametersContext.Operations, 2)
	before := slices.Clone(d.ParametersContext.Operations)

	_ = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	require.Equal(t, before, d.ParametersContext.Operations, "the alpha operation was filtered out of the caller's parameters context")
}
