package checker_test

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// deleting sunset header for a deprecated parameter is breaking
func TestBreaking_SunsetDeletedForDeprecatedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-with-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "`query` request parameter `id` sunset date deleted, but deprecated=true kept", requireChange(t, errs, checker.RequestParameterSunsetDeletedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// deleting sunset header for a deprecated parameter is breaking even if the parameter is renamed
func TestBreaking_SunsetDeletedForDeprecatedAndRenamedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-with-sunset-path.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-no-sunset-path-renamed.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "`path` request parameter `id` sunset date deleted, but deprecated=true kept", requireChange(t, errs, checker.RequestParameterSunsetDeletedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing sunset to an earlier date for a deprecated parameter with a deprecation policy is breaking
func TestBreaking_SunsetModifiedForDeprecatedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck, checker.WithDeprecation(31, 180)), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "`query` request parameter `id` sunset date changed to an earlier date, from `9999-08-10` to `2022-08-10`, new sunset date must be not earlier than `9999-08-10` and at least `180` days from now", requireChange(t, errs, checker.RequestParameterSunsetDateChangedTooSmallId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing sunset to an invalid date for a deprecated parameter is breaking
func TestBreaking_SunsetModifiedToInvalidForDeprecatedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "sunset date of the `query` request parameter `id` doesn't conform with RFC3339: `\"invalid-date\"`", requireChange(t, errs, checker.RequestParameterSunsetParseId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/param-deprecation/deprecated-invalid.yaml", errs[0].GetRevisionSource().File)
}

// changing sunset from an invalid date for a deprecated parameter is breaking
// changing an invalid sunset date of a deprecated parameter to a valid one is
// not reported, as clients had no date to rely on
func TestBreaking_SunsetModifiedFromInvalidForDeprecatedParameter(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestParameterSunsetChangedCheck, paramDeprecationFile("deprecated-invalid.yaml"), paramDeprecationFile("deprecated-future.yaml")))
}

// deleting other extension (not sunset) header for a deprecated parameter is not breaking
func TestBreaking_NonSunsetDeletedForDeprecatedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-with-other-extension.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// no change to headers for a deprecated parameter is not breaking
func TestBreaking_NoChangeToSunsetDeprecatedParameter(t *testing.T) {

	s1, err := open(paramDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-future-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// moving a parameter's sunset date earlier is breaking even when the new date
// still respects the deprecation policy
func TestRequestParameterSunsetChanged_Earlier(t *testing.T) {
	s1, err := open(paramDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(paramDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	earlier := civil.DateOf(time.Now()).AddDays(100).String()
	s2.Spec.Paths.Value("/api/test").Get.Parameters.GetByInAndName("query", "id").Extensions[diff.SunsetExtension] = toJson(t, earlier)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm)
	change := requireSingleChange(t, errs, checker.RequestParameterSunsetDateChangedEarlierId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, fmt.Sprintf("`query` request parameter `id` sunset date changed to an earlier date, from `9999-08-10` to `%s`", earlier), change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}
