package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: changing parameters of a media type is not breaking
func TestChangeMediaTypeParameters(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_params_modified.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Equal(t, "media type 'application/json' was changed to 'application/problem+json;q=1' for the response status '200'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: modifying a media type name in response to make it more specific is not breaking
func TestSpecializeMediaTypeName(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_name_modified.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Equal(t, "media type 'application/json' was changed to a more specific media type 'application/problem+json' for the response status '200'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: modifying a media type name in response to make it more general is breaking
func TestGeneralizeMediaTypeName(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_name_modified.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ERR, errs[0].GetLevel())
	require.Equal(t, "media type 'application/problem+json' was changed to a more general media type 'application/json' for the response status '200'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
