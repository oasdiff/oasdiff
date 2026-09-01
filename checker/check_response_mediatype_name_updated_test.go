package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// changing the parameters of a media type is a warning: the check cannot tell
// what the changed parameter constrains, so it cannot order the two media types
func TestChangeMediaTypeParameters(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_params_modified.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.WARN, errs[0].GetLevel())
	require.Equal(t, "media type `application/json` was changed to `application/problem+json;q=1` for the response status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.NotEmpty(t, errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// modifying a media type name in response to make it more specific is not breaking
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
	require.Equal(t, "media type `application/json` was changed to a more specific media type `application/problem+json` for the response status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// modifying a media type name in response to make it more general is breaking
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
	require.Equal(t, "media type `application/problem+json` was changed to a more general media type `application/json` for the response status `200`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// a media type parameter appearing narrows what the server may return, which
// does not break clients; the name itself is unchanged
func TestResponseMediaTypeParameterAdded(t *testing.T) {
	s1, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_param_added_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2) // the media type appears under two response statuses
	change := requireChange(t, errs, checker.ResponseMediaTypeParameterAddedId)
	require.Equal(t, checker.INFO, change.GetLevel())
	require.Equal(t, "the media type parameter `charset` was added to `application/json` for the response status `200`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// removing a media type parameter widens what the server may return, so a
// client relying on the narrower promise can break
func TestResponseMediaTypeParameterRemoved(t *testing.T) {
	s1, err := open("../data/checker/media_type_param_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/add_new_media_type_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	change := requireChange(t, errs, checker.ResponseMediaTypeParameterRemovedId)
	require.Equal(t, checker.ERR, change.GetLevel())
}

// changing a parameter's value is breaking: a client written for the old
// value, such as a charset, cannot rely on receiving it
func TestResponseMediaTypeParameterChanged(t *testing.T) {
	s1, err := open("../data/checker/media_type_param_added_revision.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/media_type_param_value_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseMediaTypeNameUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 2)
	change := requireChange(t, errs, checker.ResponseMediaTypeParameterChangedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "the media type parameter `charset` of `application/json` was changed from `utf-8` to `utf-16` for the response status `200`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}
