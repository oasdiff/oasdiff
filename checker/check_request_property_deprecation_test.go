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

func getPropertyDeprecationFile(file string) string {
	return fmt.Sprintf("../data/deprecation/%s", file)
}

// CL: detecting deprecated request properties with sunset date
func TestRequestPropertyDeprecationCheck(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedId, errs[0].GetId())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "request property 'oldField' deprecated")
}

// CL: detecting deprecated request properties in allOf schemas
func TestRequestPropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedId, errs[0].GetId())
}

// CL: ensuring no duplicate deprecation reports for the same property
func TestRequestPropertyDeprecationCheck_NoDuplicates(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
}

// BC: deprecating a property with a deprecation policy but without specifying sunset date is breaking
func TestRequestPropertyDeprecation_WithoutSunsetWithPolicy(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck).WithDeprecation(30, 100)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedSunsetMissingId, errs[0].GetId())
	require.Equal(t, "request property 'oldField' deprecated without sunset date", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a property without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestRequestPropertyDeprecation_ForAlpha(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_alpha.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset_alpha.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a property with a deprecation policy and sunset date before required deprecation period is breaking
func TestRequestPropertyDeprecation_WithEarlySunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertySunsetDateTooSmallId, errs[0].GetId())
	require.Equal(t, fmt.Sprintf("request property 'oldField' sunset date '%s' is too small, must be at least '10' days from now", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

}

// BC: deprecating a property with a deprecation policy and sunset date after required deprecation period is not breaking
func TestRequestPropertyDeprecation_WithProperSunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(10).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "request property 'oldField' deprecated")
}

// CL: properties that were re-activated
func TestRequestPropertyDeprecation_DetectsReactivated(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.RequestPropertyReactivatedId, e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "request property 'oldField' reactivated")
}

// BC: deprecating a property with an invalid sunset date format is breaking
func TestRequestPropertyDeprecation_WithInvalidSunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_invalid_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedInvalidId, errs[0].GetId())
}

// CL: deprecating a request property with invalid stability level is skipped (handled in CheckBackwardCompatibility)
func TestRequestPropertyDeprecation_WithInvalidStability(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	// Set invalid stability level on the operation
	s2.Spec.Paths.Value("/test").Post.Extensions[diff.XStabilityLevelExtension] = toJson(t, "invalid-stability")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.RequestPropertyDeprecationCheck(d, osm, checker.NewConfig(nil))
	require.Empty(t, changes)
}

// CL: message has no details when request property deprecated without sunset or stability
func TestRequestPropertyDeprecation_MessageWithoutDetails(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset_no_stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, "request property 'oldField' deprecated", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
