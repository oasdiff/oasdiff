package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding 'allOf' subschema to the request body or request body property
func TestRequestPropertyAllOfAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAllOfAddedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Level:       checker.ERR,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_added_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAllOfAddedId,
			Args:        []any{"#/components/schemas/Breed3", "/allOf[#/components/schemas/Dog]/breed"},
			Level:       checker.ERR,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_added_revision.yaml"),
			OperationId: "updatePets",
		}}, errs)
}

// CL: adding 'allOf' subschema ($ref) with source tracking verifies source points to component definition
func TestRequestPropertyAllOfAdded_WithSources_Ref(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/request_property_all_of_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	for _, err := range errs {
		if err.GetId() == checker.RequestBodyAllOfAddedId {
			// Added $ref subschema: revision source should point to the component definition (Rabbit), not the allOf keyword
			require.NotEmpty(t, err.GetRevisionSource())
			require.Equal(t, "../data/checker/request_property_all_of_added_revision.yaml", err.GetRevisionSource().File)
			require.NotZero(t, err.GetRevisionSource().Line)
		}
	}
}

// CL: adding inline 'allOf' subschema with source tracking verifies source points to inline subschema
func TestRequestPropertyAllOfAdded_WithSources_Inline(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/request_property_all_of_inline_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_inline_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestBodyAllOfAddedId, errs[0].GetId())

	// Added inline subschema: revision source should point to the specific subschema, not the allOf keyword
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/checker/request_property_all_of_inline_added_revision.yaml", errs[0].GetRevisionSource().File)
	// The added subschema is at line 23 ("- type: object" for prop3), not line 14 (allOf keyword)
	require.Greater(t, errs[0].GetRevisionSource().Line, 14)
}

// CL: removing 'allOf' subschema from the request body or request body property
func TestRequestPropertyAllOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Level:       checker.WARN,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_removed_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Breed3", "/allOf[#/components/schemas/Dog]/breed"},
			Level:       checker.WARN,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_removed_revision.yaml"),
			OperationId: "updatePets",
		}}, errs)
}
