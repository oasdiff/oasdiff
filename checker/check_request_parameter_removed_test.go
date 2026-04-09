package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: deleting a parameter without deprecation is breaking
func TestBreaking_DeletedParameter(t *testing.T) {
	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterRemovedId, errs[0].GetId())
	require.Equal(t, "deleted the `query` request parameter `id`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "This is a warning because some apps may return an error when receiving a parameter that they do not expect. It is recommended to deprecate the parameter first.", errs[0].GetComment(checker.NewDefaultLocalizer()))
}

// BC: deleting a parameter after sunset date is not breaking
func TestBreaking_ParameterDeprecationPast(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deleting a parameter before sunset date is breaking
func TestBreaking_ParameterDeprecationFuture(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ParameterRemovedBeforeSunsetId, errs[0].GetId())
	require.Equal(t, "deleted the `query` request parameter `id` before the sunset date `9999-08-10`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deleting a deprecated parameter without sunset date is not breaking
func TestBreaking_ParameterDeprecationNoSunset(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: removing a parameter without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestBreaking_RemovedParameterForAlpha(t *testing.T) {
	s1, err := open(getParameterDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: removing a deprecated parameter with an invalid date is breaking
func TestBreaking_RemoveParameterWithInvalidSunset(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date for the `query` request parameter `id`: `sunset date doesn't conform with RFC3339: invalid-date`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/param-deprecation/sunset.yaml", errs[0].GetSource())
}
