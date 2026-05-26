package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// CL: adding 'anyOf' schema to the request body or request body property
func TestRequestPropertyAnyOfAdded(t *testing.T) {
	s1, err := open("../data/checker/request_property_any_of_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_any_of_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAnyOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAnyOfAddedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Level:       checker.INFO,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_any_of_added_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAnyOfAddedId,
			Args:        []any{"#/components/schemas/Breed3", "anyOf[#/components/schemas/Dog]/breed"},
			Level:       checker.INFO,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_any_of_added_revision.yaml"),
			OperationId: "updatePets",
		}}, errs)
}

// CL: adding 'anyOf' subschema ($ref) with source tracking
func TestRequestPropertyAnyOfAdded_WithSources(t *testing.T) {
	loader := newLoaderWithOriginTracking()
	s1, err := open("../data/checker/request_property_any_of_added_base.yaml", loader)
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_any_of_added_revision.yaml", loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAnyOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	for _, err := range errs {
		if err.GetId() == checker.RequestBodyAnyOfAddedId {
			require.NotEmpty(t, err.GetRevisionSource())
			require.Equal(t, "../data/checker/request_property_any_of_added_revision.yaml", err.GetRevisionSource().File)
			require.NotZero(t, err.GetRevisionSource().Line)
		}
	}
}

// CL: removing 'anyOf' schema from the request body or request body property
func TestRequestPropertyAnyOfRemoved(t *testing.T) {
	s1, err := open("../data/checker/request_property_any_of_removed_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_any_of_removed_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAnyOfUpdatedCheck), d, osm, checker.INFO)

	require.Len(t, errs, 2)

	require.ElementsMatch(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAnyOfRemovedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Level:       checker.ERR,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_any_of_removed_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAnyOfRemovedId,
			Args:        []any{"#/components/schemas/Breed3", "anyOf[#/components/schemas/Dog]/breed"},
			Level:       checker.ERR,
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_any_of_removed_revision.yaml"),
			OperationId: "updatePets",
		}}, errs)
}

// BC: refactoring an anyOf branch from inline enum to an equivalent $ref is not breaking
func TestRequestPropertyAnyOfInlineEnumRefactorToRef(t *testing.T) {
	s1, err := open("../data/checker/request_property_any_of_ref_inline_enum_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_any_of_ref_inline_enum_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	breakingChanges := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, breakingChanges)

	anyOfChanges := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAnyOfUpdatedCheck), d, osm, checker.INFO)
	require.False(t, containsId(anyOfChanges, checker.RequestPropertyAnyOfAddedId))
	require.False(t, containsId(anyOfChanges, checker.RequestPropertyAnyOfRemovedId))
	require.Empty(t, anyOfChanges)
}

// CL: body-level anyOf added/removed changes carry media-type details so
// reports can tell which media type each change belongs to when a request
// body has more than one. Without WithDetails the entries are identical
// and visitors can't tell apart "added Rabbit on application/json" from
// "added Rabbit on application/xml".
func TestRequestBodyAnyOfMultiMediaTypeDetails(t *testing.T) {
	s1, err := open("../data/checker/request_body_any_of_media_type_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_body_any_of_media_type_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAnyOfUpdatedCheck), d, osm, checker.INFO)

	var addedDetails, removedDetails []string
	for _, e := range errs {
		c := e.(checker.ApiChange)
		switch c.Id {
		case checker.RequestBodyAnyOfAddedId:
			addedDetails = append(addedDetails, c.Details)
		case checker.RequestBodyAnyOfRemovedId:
			removedDetails = append(removedDetails, c.Details)
		}
	}

	require.ElementsMatch(t, []string{
		"(media type: application/json)",
		"(media type: application/xml)",
	}, addedDetails)
	require.ElementsMatch(t, []string{
		"(media type: application/json)",
		"(media type: application/xml)",
	}, removedDetails)
}

// CL: no changes when paths diff is nil
func TestRequestPropertyAnyOfNoPathsDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyAnyOfUpdatedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when operations diff is nil
func TestRequestPropertyAnyOfNoOperationsDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyAnyOfUpdatedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when request body diff is nil
func TestRequestPropertyAnyOfNoRequestBodyDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{
							"POST": &diff.MethodDiff{},
						},
					},
				},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyAnyOfUpdatedCheck(d, osm, config)
	require.Len(t, errs, 0)
}

// CL: no changes when schema diff is nil
func TestRequestPropertyAnyOfNoSchemaDiff(t *testing.T) {
	config := &checker.Config{}
	d := &diff.Diff{
		PathsDiff: &diff.PathsDiff{
			Modified: diff.ModifiedPaths{
				"/test": &diff.PathDiff{
					OperationsDiff: &diff.OperationsDiff{
						Modified: diff.ModifiedOperations{
							"POST": &diff.MethodDiff{
								RequestBodyDiff: &diff.RequestBodyDiff{
									ContentDiff: &diff.ContentDiff{
										MediaTypeModified: diff.ModifiedMediaTypes{
											"application/json": &diff.MediaTypeDiff{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	osm := &diff.OperationsSourcesMap{}

	errs := checker.RequestPropertyAnyOfUpdatedCheck(d, osm, config)
	require.Len(t, errs, 0)
}
