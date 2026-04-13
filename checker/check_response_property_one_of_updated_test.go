package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// BC: adding 'oneOf' schema to the response body or response body property is breaking
func TestResponsePropertyOneOfAdded(t *testing.T) {
	s1, err := open("../data/checker/response_property_one_of_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyOneOfUpdated), d, osm, checker.INFO)

	require.Len(t, errs, 3)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.ResponseBodyOneOfAddedId,
			Args:        []any{"#/components/schemas/Rabbit", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_added_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyOneOfAddedId,
			Args:        []any{"#/components/schemas/Breed3", "/oneOf[#/components/schemas/Dog]/breed", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_added_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyOneOfAddedId,
			Args:        []any{"subschema #2: Dark brown types", "/oneOf[#/components/schemas/Fox]/breed", "200"},
			Level:       checker.ERR,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_added_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}

// CL: adding 'oneOf' subschema ($ref) with source tracking
func TestResponsePropertyOneOfAdded_WithSources(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/response_property_one_of_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyOneOfUpdated), d, osm, checker.INFO)

	require.Len(t, errs, 3)

	for _, err := range errs {
		if err.GetId() == checker.ResponseBodyOneOfAddedId {
			require.NotEmpty(t, err.GetRevisionSource())
			require.Equal(t, "../data/checker/response_property_one_of_added_revision.yaml", err.GetRevisionSource().File)
			require.NotZero(t, err.GetRevisionSource().Line)
		}
	}
}

// CL: removing 'oneOf' schema from the response body or response body property
func TestResponsePropertyOneOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/response_property_one_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/response_property_one_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyOneOfUpdated), d, osm, checker.INFO)

	require.Len(t, errs, 3)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.ResponseBodyOneOfRemovedId,
			Args:        []any{"#/components/schemas/Rabbit", "200"},
			Level:       checker.INFO,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_removed_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyOneOfRemovedId,
			Args:        []any{"#/components/schemas/Breed3", "/oneOf[#/components/schemas/Dog]/breed", "200"},
			Level:       checker.INFO,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_removed_revision.yaml"),
			OperationId: "listPets",
		},
		{
			Id:          checker.ResponsePropertyOneOfRemovedId,
			Args:        []any{"subschema #2: Dark brown types", "/oneOf[#/components/schemas/Fox]/breed", "200"},
			Level:       checker.INFO,
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/response_property_one_of_removed_revision.yaml"),
			OperationId: "listPets",
		}}, errs)
}
