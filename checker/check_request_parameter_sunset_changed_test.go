package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: deleting sunset header for a deprecated parameter is breaking
func TestBreaking_SunsetDeletedForDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-with-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetDeletedId, errs[0].GetId())
	require.Equal(t, "`query` request parameter `id` sunset date deleted, but deprecated=true kept", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deleting sunset header for a deprecated parameter is breaking even if the parameter is renamed
func TestBreaking_SunsetDeletedForDeprecatedAndRenamedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-with-sunset-path.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset-path-renamed.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetDeletedId, errs[0].GetId())
	require.Equal(t, "`path` request parameter `id` sunset date deleted, but deprecated=true kept", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing sunset to an earlier date for a deprecated parameter with a deprecation policy is breaking
func TestBreaking_SunsetModifiedForDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck).WithDeprecation(31, 180), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetDateChangedTooSmallId, errs[0].GetId())
	require.Equal(t, "`query` request parameter `id` sunset date changed to an earlier date, from `9999-08-10` to `2022-08-10`, new sunset date must be not earlier than `9999-08-10` and at least `180` days from now", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing sunset to an invalid date for a deprecated parameter is breaking
func TestBreaking_SunsetModifiedToInvalidForDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date for the `query` request parameter `id`: `sunset date doesn't conform with RFC3339: invalid-date`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/param-deprecation/deprecated-invalid.yaml", errs[0].GetSource())
}

// BC: changing sunset from an invalid date for a deprecated parameter is breaking
func TestBreaking_SunsetModifiedFromInvalidForDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date for the `query` request parameter `id`: `sunset date doesn't conform with RFC3339: invalid-date`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/param-deprecation/deprecated-invalid.yaml", errs[0].GetSource())
}

// BC: deleting other extension (not sunset) header for a deprecated parameter is not breaking
func TestBreaking_NonSunsetDeletedForDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-with-other-extension.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: no change to headers for a deprecated parameter is not breaking
func TestBreaking_NoChangeToSunsetDeprecatedParameter(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-future-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterSunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}
