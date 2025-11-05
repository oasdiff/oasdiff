package checker_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func open(file string) (*load.SpecInfo, error) {
	return load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(file))
}

func getDeprecationFile(file string) string {
	return fmt.Sprintf("../data/deprecation/%s", file)
}

func singleCheckConfig(c checker.BackwardCompatibilityCheck) *checker.Config {
	return checker.NewConfig(checker.BackwardCompatibilityChecks{c}).WithSingleCheck(c)
}

func allChecksConfig() *checker.Config {
	return checker.NewConfig(checker.GetAllChecks())
}

// BC: deprecating an operation with a deprecation policy and an invalid sunset date is breaking
func TestBreaking_DeprecationWithInvalidSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-with-invalid-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIDeprecatedSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date: 'sunset date doesn't conform with RFC3339: invalid'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating an operation with a deprecation policy and an invalid stability level is breaking
func TestBreaking_DeprecationWithInvalidStabilityLevel(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-with-invalid-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: 'value is not one of draft, alpha, beta or stable: \"invalid\"'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-with-invalid-stability.yaml", errs[0].GetSource())
}

// BC: deprecating an operation without a deprecation policy but without specifying sunset date is not breaking
func TestBreaking_DeprecationWithoutSunsetNoPolicy(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(0, 0)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Empty(t, errs)
}

// BC: deprecating an operation with a deprecation policy but without specifying sunset date is breaking
func TestBreaking_DeprecationWithoutSunsetWithPolicy(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(30, 100)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIDeprecatedSunsetMissingId, errs[0].GetId())
	require.Equal(t, "sunset date is missing for deprecated API", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating an operation with a default deprecation policy but without specifying sunset date is not breaking
func TestBreaking_DeprecationWithoutSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Empty(t, errs)
}

// BC: deprecating an operation without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestBreaking_DeprecationForAlpha(t *testing.T) {

	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-no-sunset-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating an operation without a deprecation policy and without specifying sunset date is not breaking for draft level
func TestBreaking_DeprecationForDraft(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)
	draft := toJson(t, checker.STABILITY_DRAFT)
	s1.Spec.Paths.Value("/api/test").Get.Extensions["x-stability-level"] = draft

	s2, err := open(getDeprecationFile("deprecated-no-sunset-alpha-stability.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/test").Get.Extensions["x-stability-level"] = draft

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

func toJson(t *testing.T, value string) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}

// BC: deprecating an operation with a deprecation policy and sunset date before required deprecation period is breaking
func TestBreaking_DeprecationWithEarlySunset(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APISunsetDateTooSmallId, errs[0].GetId())
	require.Equal(t, fmt.Sprintf("sunset date '%s' is too small, must be at least '10' days from now", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating an operation with a deprecation policy and sunset date after required deprecation period is not breaking
func TestBreaking_DeprecationWithProperSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, civil.DateOf(time.Now()).AddDays(10).String())

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	c := singleCheckConfig(checker.APIDeprecationCheck).WithDeprecation(0, 10)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	// only a non-breaking change detected
	require.Equal(t, checker.EndpointDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "endpoint deprecated")
}

// CL: path operations that became deprecated
func TestApiDeprecated_DetectsDeprecatedOperations(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIDeprecationCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.EndpointDeprecatedId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "endpoint deprecated")
}

// CL: path operations that were re-activated
func TestApiDeprecated_DetectsReactivatedOperations(t *testing.T) {
	s1, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIDeprecationCheck), d, osm, checker.INFO)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.EndpointReactivatedId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "endpoint reactivated")
}

func TestBreaking_InvaidStability(t *testing.T) {

	s1, err := open(getDeprecationFile("invalid-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIDeprecationCheck), d, osm)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.APIInvalidStabilityLevelId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Equal(t, "failed to parse stability level: 'value is not one of draft, alpha, beta or stable: \"ga\"'", e0.GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/invalid-stability.yaml", errs[0].GetSource())
}
