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

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAllOfAddedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_added_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAllOfAddedId,
			Args:        []any{"#/components/schemas/Breed3", "allOf[#/components/schemas/Dog]/breed"},
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

	requireSingleChange(t, errs, checker.RequestBodyAllOfAddedId)

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

	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestBodyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Rabbit"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_removed_revision.yaml"),
			OperationId: "updatePets",
		},
		{
			Id:          checker.RequestPropertyAllOfRemovedId,
			Args:        []any{"#/components/schemas/Breed3", "allOf[#/components/schemas/Dog]/breed"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource("../data/checker/request_property_all_of_removed_revision.yaml"),
			OperationId: "updatePets",
		}}, errs)
}

// CL: adding an allOf subschema whose body is annotation-only (title,
// description, examples, default, externalDocs, $comment) is a wire-
// contract no-op. We don't emit it at the original ERR severity (that
// would be a false positive on the breaking-change view), but we do
// emit it at INFO so the document-level change stays auditable in
// `oasdiff changelog`. See OAS discussion
// https://github.com/OAI/OpenAPI-Specification/discussions/3793
// (handrews: "if you add an allOf that only includes annotation
// keywords (like title), that's not really a breaking change the way
// it is if you add another constraint that invalidates previously-
// valid instances.")
func TestRequestPropertyAllOfAdded_AnnotationOnly_EmitsInfo(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_annotation_only_added_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_annotation_only_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	requireSingleChange(t, errs, checker.RequestBodyAllOfAddedAnnotationOnlyId)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	// And critically: not the original ERR-level breaking-change ID.
	require.NotEqual(t, checker.RequestBodyAllOfAddedId, errs[0].GetId())
}

// CL: same as above but the annotation-only allOf lives on a nested
// property's schema, not on the body. Covers the info.walkProperties
// code path; the body-level test only exercises the outer walker.
func TestRequestPropertyAllOfAdded_AnnotationOnly_AtProperty_EmitsInfo(t *testing.T) {
	s1, err := open("../data/checker/request_property_all_of_annotation_only_added_at_property_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/request_property_all_of_annotation_only_added_at_property_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyAllOfUpdatedCheck), d, osm, checker.INFO)

	requireSingleChange(t, errs, checker.RequestPropertyAllOfAddedAnnotationOnlyId)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.NotEqual(t, checker.RequestPropertyAllOfAddedId, errs[0].GetId())
}
