package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: deleting sunset header for a deprecated endpoint is breaking
func TestBreaking_SunsetDeletedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-with-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APISunsetDeletedId, errs[0].GetId())
	require.Equal(t, "api sunset date deleted, but deprecated=true kept", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing sunset to an earlier date for a deprecated endpoint with a deprecation policy is breaking
func TestBreaking_SunsetModifiedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck).WithDeprecation(31, 180), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APISunsetDateChangedTooSmallId, errs[0].GetId())
	require.Equal(t, "api sunset date changed to an earlier date, from `9999-08-10` to `2022-08-10`, new sunset date must be not earlier than `9999-08-10` and at least `180` days from now", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing sunset to an invalid date for a deprecated endpoint is breaking
func TestBreaking_SunsetModifiedToInvalidForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date: `sunset date doesn't conform with RFC3339: invalid-date`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: changing sunset from an invalid date for a deprecated endpoint is breaking
func TestBreaking_SunsetModifiedFromInvalidForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date: `sunset date doesn't conform with RFC3339: invalid-date`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-invalid.yaml", errs[0].GetSource())
}

// BC: deleting other extension (not sunset) header for a deprecated endpoint is not breaking
func TestBreaking_NonSunsetDeletedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-with-other-extension.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: no change to headers for a deprecated endpoint is not breaking
func TestBreaking_NoChangeToSunsetDeprecatedEndpoint(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-future-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}
