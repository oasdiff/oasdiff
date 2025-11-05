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

func getParameterDeprecationFile(file string) string {
	return fmt.Sprintf("../data/param-deprecation/%s", file)
}

// BC: deprecating a parameter with a deprecation policy and an invalid sunset date is breaking
func TestBreaking_ParameterDeprecationWithInvalidSunset(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date for the 'query' request parameter 'id': 'sunset date doesn't conform with RFC3339: invalid-date'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a parameter without a deprecation policy but without specifying sunset date is not breaking
func TestBreaking_ParameterDeprecationWithoutSunsetNoPolicy(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck).WithDeprecation(0, 0)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a parameter with a deprecation policy but without specifying sunset date is breaking
func TestBreaking_ParameterDeprecationWithoutSunsetWithPolicy(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck).WithDeprecation(30, 100)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterDeprecatedSunsetMissingId, errs[0].GetId())
	require.Equal(t, "'query' request parameter 'id' was deprecated without sunset date", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a parameter with a default deprecation policy but without specifying sunset date is not breaking
func TestBreaking_ParameterDeprecationWithoutSunset(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Empty(t, errs)
}

// BC: deprecating an operation without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestBreaking_ParameterDeprecationForAlpha(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-no-sunset-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a parameter with a deprecation policy and sunset date before required deprecation period is breaking
func TestBreaking_ParameterDeprecationWithEarlySunset(t *testing.T) {
	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()

	s2.Spec.Paths.Value("/api/test").GetOperation("GET").Parameters.GetByInAndName("query", "id").Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterSunsetDateTooSmallId, errs[0].GetId())
	require.Equal(t, fmt.Sprintf("'query' request parameter 'id' sunset date '%s' is too small, must be at least '10' days from now", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a parameter with a deprecation policy and sunset date after required deprecation period is not breaking
func TestBreaking_ParameterDeprecationWithProperSunset(t *testing.T) {

	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/test").GetOperation("GET").Parameters.GetByInAndName("query", "id").Extensions[diff.SunsetExtension] = toJson(t, civil.DateOf(time.Now()).AddDays(10).String())

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	c := singleCheckConfig(checker.RequestParameterDeprecationCheck).WithDeprecation(0, 10)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	// only a non-breaking change detected
	require.Equal(t, checker.RequestParameterDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "'query' request parameter 'id' was deprecated")
}

// CL: parameters that became deprecated
func TestParameterDeprecated_DetectsDeprecated(t *testing.T) {
	s1, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterDeprecationCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.RequestParameterDeprecatedId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "'query' request parameter 'id' was deprecated")
}

// CL: parameters that were re-activated
func TestParameterDeprecated_DetectsReactivated(t *testing.T) {
	s1, err := open(getParameterDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getParameterDeprecationFile("base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterDeprecationCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.RequestParameterReactivatedId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "'query' request parameter 'id' was reactivated")
}
