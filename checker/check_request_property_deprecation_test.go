package checker_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// detecting deprecated request properties with sunset date
func TestRequestPropertyDeprecationCheck(t *testing.T) {
	s1, err := open(deprecationFile("request_property_deprecation_base.yaml"))
	require.NoError(t, err)
	s2, err := open(deprecationFile("request_property_deprecation_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedWithSunsetId)
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "request property `oldField` deprecated")
}

// detecting deprecated request properties in allOf schemas with multiple media types
func TestRequestPropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open(deprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(deprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	// With multiple media types (json and xml), we get one report per media type with distinct details
	require.Len(t, errs, 2)
	requireChange(t, errs, checker.RequestPropertyDeprecatedWithSunsetId)
	requireChange(t, errs, checker.RequestPropertyDeprecatedWithSunsetId)
	// Each message should include media type context
	msg0 := errs[0].GetUncolorizedText(checker.NewDefaultLocalizer())
	msg1 := errs[1].GetUncolorizedText(checker.NewDefaultLocalizer())
	require.Contains(t, msg0, "media type:")
	require.Contains(t, msg1, "media type:")
	// Messages should be distinct (different media types)
	require.NotEqual(t, msg0, msg1)
}

// each media type gets its own report with distinct details (issue #594)
func TestRequestPropertyDeprecationCheck_MediaTypeContext(t *testing.T) {
	s1, err := open(deprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(deprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
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

// deprecating a property with a deprecation policy but without specifying sunset date is breaking
func TestRequestPropertyDeprecation_WithoutSunsetWithPolicy(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_no_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck, checker.WithDeprecation(30, 100))
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedSunsetMissingId)
	require.Equal(t, "request property `oldField` deprecated without sunset date", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// deprecating a property without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestRequestPropertyDeprecation_ForAlpha(t *testing.T) {
	s1, err := open(deprecationFile("property_base_alpha.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_no_sunset_alpha.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm)
	require.Empty(t, errs)
}

// deprecating a property with a deprecation policy and sunset date before required deprecation period is breaking
func TestRequestPropertyDeprecation_WithEarlySunset(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck, checker.WithDeprecation(0, 10))
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	requireSingleChange(t, errs, checker.RequestPropertySunsetDateTooSmallId)
	require.Equal(t, fmt.Sprintf("request property `oldField` sunset date `%s` is too small, must be at least `10` days from now", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))

}

// deprecating a property with a deprecation policy and sunset date after required deprecation period is not breaking
func TestRequestPropertyDeprecation_WithProperSunset(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(10).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck, checker.WithDeprecation(0, 10))
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedWithSunsetId)
	require.Equal(t, checker.INFO, errs[0].GetLevel())
	require.Contains(t, errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()), "request property `oldField` deprecated")
}

// properties that were re-activated
func TestRequestPropertyDeprecation_DetectsReactivated(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_base_stable.yaml"))
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
	require.Contains(t, e0.GetUncolorizedText(checker.NewDefaultLocalizer()), "request property `oldField` reactivated")
}

// deprecating a property with an invalid sunset date format is breaking
func TestRequestPropertyDeprecation_WithInvalidSunset(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_invalid_sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedInvalidId)
}

// deprecating a request property with invalid stability level is skipped (handled in CheckBackwardCompatibility)
func TestRequestPropertyDeprecation_WithInvalidStability(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	// Set invalid stability level on the operation
	s2.Spec.Paths.Value("/test").Post.Extensions[diff.XStabilityLevelExtension] = toJson(t, "invalid-stability")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	changes := checker.RequestPropertyDeprecationCheck(d, osm, checker.NewConfig(nil))
	require.Empty(t, changes)
}

// message has no details when request property deprecated without sunset or stability
func TestRequestPropertyDeprecation_MessageWithoutDetails(t *testing.T) {
	s1, err := open(deprecationFile("property_base.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_no_sunset_no_stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedId)
	require.Equal(t, "request property `oldField` deprecated", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// message includes sunset date when request property deprecated with valid sunset
func TestRequestPropertyDeprecation_MessageWithSunsetDate(t *testing.T) {
	s1, err := open(deprecationFile("property_base_stable.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	sunsetDate := civil.DateOf(time.Now()).AddDays(30).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck, checker.WithDeprecation(0, 10))
	errs := checker.CheckBackwardCompatibilityUntilLevel(c, d, osm, checker.INFO)
	requireSingleChange(t, errs, checker.RequestPropertyDeprecatedWithSunsetId)
	require.Equal(t, fmt.Sprintf("request property `oldField` deprecated with sunset date `%s` (stability: stable)", sunsetDate), errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// source location points to the deprecated field, not the operation level
func TestRequestPropertyDeprecationCheck_SourceLocation(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	s1, err := load.NewSpecInfo(loader, load.NewSource(deprecationFile("request_property_deprecation_base.yaml")))
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(loader, load.NewSource(deprecationFile("request_property_deprecation_spec.yaml")))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertyDeprecationCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)

	// RevisionSource must point to the `deprecated: true` line, not the operation level
	revSource := errs[0].GetRevisionSource()
	require.NotNil(t, revSource, "revision source must be set")
	require.Equal(t, 37, revSource.Line, "source must point to `deprecated: true` line, not operation level")
}
