package checker_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CL: detecting deprecated response properties with sunset date
func TestResponsePropertyDeprecationCheck(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[0].GetId())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "response property 'legacyField' deprecated")
}

// CL: detecting deprecated response properties in allOf schemas with multiple media types
func TestResponsePropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm, checker.INFO)
	// With multiple media types (json and xml), we get one report per media type with distinct details
	require.Len(t, errs, 2)
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[1].GetId())
	// Each message should include media type context
	msg0 := errs[0].GetUncolorizedText(checker.NewDefaultLocalizer())
	msg1 := errs[1].GetUncolorizedText(checker.NewDefaultLocalizer())
	require.Contains(t, msg0, "media type:")
	require.Contains(t, msg1, "media type:")
	// Messages should be distinct (different media types)
	require.NotEqual(t, msg0, msg1)
}

// CL: each media type gets its own report with distinct details (issue #594)
func TestResponsePropertyDeprecationCheck_MediaTypeContext(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm, checker.INFO)
	// Multiple media types should result in multiple reports with media type context
	require.Len(t, errs, 2)
	mediaTypes := make(map[string]bool)
	for _, err := range errs {
		msg := err.GetUncolorizedText(checker.NewDefaultLocalizer())
		if strings.Contains(msg, "application/json") {
			mediaTypes["application/json"] = true
		}
		if strings.Contains(msg, "application/xml") {
			mediaTypes["application/xml"] = true
		}
	}
	require.True(t, mediaTypes["application/json"], "should have application/json media type")
	require.True(t, mediaTypes["application/xml"], "should have application/xml media type")
}

// BC: deprecating a response property with a deprecation policy but without specifying sunset date is breaking
func TestResponsePropertyDeprecation_WithoutSunsetWithPolicy(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.ResponsePropertyDeprecationCheck).WithDeprecation(30, 100)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyDeprecatedSunsetMissingId, errs[0].GetId())
	require.Equal(t, "response property 'legacyField' deprecated without sunset date", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a response property without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestResponsePropertyDeprecation_ForAlpha(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_alpha.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset_alpha.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deprecating a response property with a deprecation policy and sunset date before required deprecation period is breaking
func TestResponsePropertyDeprecation_WithEarlySunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Components.Schemas["TestResponse"].Value.Properties["legacyField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.ResponsePropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertySunsetDateTooSmallId, errs[0].GetId())
	require.Equal(t, fmt.Sprintf("response property 'legacyField' sunset date '%s' is too small, must be at least '10' days from now", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: deprecating a response property with a deprecation policy and sunset date after required deprecation period is not breaking
func TestResponsePropertyDeprecation_WithProperSunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(10).String()
	s2.Spec.Components.Schemas["TestResponse"].Value.Properties["legacyField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.ResponsePropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "response property 'legacyField' deprecated")
}

// CL: response properties that were re-activated
func TestResponsePropertyDeprecation_DetectsReactivated(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_deprecated.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.ResponsePropertyReactivatedId, e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/test", e0.Path)
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "response property 'legacyField' reactivated")
}

// BC: deprecating a response property with an invalid sunset date format is breaking
func TestResponsePropertyDeprecation_WithInvalidSunset(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_invalid_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.ResponsePropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyDeprecatedInvalidId, errs[0].GetId())
}

// CL: deprecating a response property with invalid stability level is skipped (handled in CheckBackwardCompatibility)
func TestResponsePropertyDeprecation_WithInvalidStability(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	// Set invalid stability level on the operation
	s2.Spec.Paths.Value("/test").Post.Extensions[diff.XStabilityLevelExtension] = toJson(t, "invalid-stability")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.ResponsePropertyDeprecationCheck(d, osm, checker.NewConfig(nil))
	require.Empty(t, changes)
}

// CL: message has no details when response property deprecated without sunset or stability
func TestResponsePropertyDeprecation_MessageWithoutDetails(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("property_base.yaml"))
	require.NoError(t, err)

	s2, err := open(getPropertyDeprecationFile("property_deprecated_no_sunset_no_stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponsePropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, "response property 'legacyField' deprecated", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}
