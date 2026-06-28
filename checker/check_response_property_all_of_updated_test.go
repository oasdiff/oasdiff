package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// adding 'allOf' subschema to the response body or response body property
func TestResponsePropertyAllOfAdded(t *testing.T) {
	s1, err := open("../data/checker/response_property_all_of_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_all_of_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.ResponseBodyAllOfAddedId,
			Args:        []any{"#/components/schemas/Rabbit", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_all_of_added_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyAllOfAddedId,
			Args:        []any{"#/components/schemas/Breed3", "allOf[#/components/schemas/Dog]/breed", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_all_of_added_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}

// adding 'allOf' subschema ($ref) with source tracking
func TestResponsePropertyAllOfAdded_WithSources_Ref(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/response_property_all_of_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_all_of_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	for _, err := range errs {
		if err.GetId() == checker.ResponseBodyAllOfAddedId {
			require.NotEmpty(t, err.GetRevisionSource())
			require.Equal(t, "../data/checker/response_property_all_of_added_revision.yaml", err.GetRevisionSource().File)
			require.NotZero(t, err.GetRevisionSource().Line)
		}
	}
}

// adding inline 'allOf' subschema to the response body with source tracking
func TestResponsePropertyAllOfAdded_WithSources_Inline(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/response_property_all_of_inline_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_all_of_inline_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	requireSingleChange(t, errs, checker.ResponseBodyAllOfAddedId)

	// Added inline subschema: source should point to specific subschema, not allOf keyword
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/response_property_all_of_inline_added_revision.yaml", errs[0].GetRevisionSource().File)
	// The allOf keyword is at line 17; the added subschema is further down
	require.Greater(t, errs[0].GetRevisionSource().Line, 17)
}

// removing 'allOf' subschema from the response body or response body property
func TestResponsePropertyAllOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_all_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_all_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.ResponseBodyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Rabbit", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_all_of_removed_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Breed3", "allOf[#/components/schemas/Dog]/breed", "200"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_all_of_removed_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}

// adding an allOf subschema to a response body whose body is
// annotation-only is emitted as a distinct INFO change instead of the
// original allOf-added INFO. The response-side severity is unchanged
// (INFO either way), but the new ID makes the wire-contract-neutral
// nature of the change explicit in tooling and audit trails. Mirrors
// the request-side test; see OAS discussion #3793 for the motivating
// case.
func TestResponsePropertyAllOfAdded_AnnotationOnly_EmitsInfoWithDistinctId(t *testing.T) {
	s1, err := open("../data/checker/response_property_all_of_annotation_only_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_all_of_annotation_only_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	requireSingleChange(t, errs, checker.ResponseBodyAllOfAddedAnnotationOnlyId)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	// Distinct from the original allOf-added INFO ID.
	require.NotEqual(t, checker.ResponseBodyAllOfAddedId, errs[0].GetId())
}
