package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func loadOneOfDescriptionSpecs(t *testing.T) (*openapi3.T, *openapi3.T) {
	t.Helper()

	loader := openapi3.NewLoader()
	s1, err := loader.LoadFromFile("../data/checker/response_property_one_of_description_base.yaml")
	require.NoError(t, err)
	s2, err := loader.LoadFromFile("../data/checker/response_property_one_of_description_revision.yaml")
	require.NoError(t, err)

	return s1, s2
}

func oneOfDiff(t *testing.T, d *diff.Diff) *diff.SubschemasDiff {
	t.Helper()

	return d.PathsDiff.Modified["/pets"].OperationsDiff.Modified["GET"].ResponsesDiff.
		Modified["200"].ContentDiff.MediaTypeModified["application/json"].SchemaDiff.
		PropertiesDiff.Modified["result"].OneOfDiff
}

// Two untitled oneOf branches whose item schemas differ only in a property
// description are the same two subschemas, so each is paired with its
// counterpart and the description change is reported inside it (oasdiff #1403).
func TestSubschemas_AnnotationOnlyPairs(t *testing.T) {
	s1, s2 := loadOneOfDescriptionSpecs(t)

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	oneOf := oneOfDiff(t, d)
	require.Empty(t, oneOf.Added)
	require.Empty(t, oneOf.Deleted)
	require.Len(t, oneOf.Modified, 2)
	for _, modified := range oneOf.Modified {
		require.Equal(t, modified.Base.Index, modified.Revision.Index)
		require.Equal(t,
			&diff.ValueDiff{From: "The shelter that registered the pet.", To: "The shelter that registered the pet, if any."},
			modified.Diff.ItemsDiff.PropertiesDiff.Modified["shelter_id"].DescriptionDiff)
	}
}

// The pairing is by validation contract, so branches that differ in what they
// accept are still reported as removed and added.
func TestSubschemas_ValidationChangeDoesNotPair(t *testing.T) {
	s1, s2 := loadOneOfDescriptionSpecs(t)

	// Both branches now accept objects instead of arrays, which no annotation
	// comparison can look past.
	for _, subschema := range s2.Paths.Value("/pets").Get.Responses.Value("200").
		Value.Content["application/json"].Schema.Value.Properties["result"].Value.OneOf {
		subschema.Value.Type = &openapi3.Types{"object"}
		subschema.Value.Items = nil
	}

	d, err := diff.Get(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	oneOf := oneOfDiff(t, d)
	require.Len(t, oneOf.Added, 2)
	require.Len(t, oneOf.Deleted, 2)
	require.Empty(t, oneOf.Modified)
}
