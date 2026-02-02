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

// detecting deprecated request properties with sunset date
func TestRequestPropertyDeprecationCheck(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated request property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected RequestPropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// detecting deprecated request properties in allOf schemas
func TestRequestPropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated request allOf property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected RequestPropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// Ensuring no duplicate deprecation reports for the same property
func TestRequestPropertyDeprecationCheck_NoDuplicates(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("request_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	// Count occurrences of each property
	propCount := make(map[string]int)
	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyDeprecatedId {
			propCount[c.GetText(checker.NewDefaultLocalizer())]++
		}
	}

	// Each property should only appear once
	for prop, count := range propCount {
		if count > 1 {
			t.Errorf("Property %s appears %d times, expected 1", prop, count)
		}
	}
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

	// Update the sunset date in the spec to be too small
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.RequestPropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertySunsetDateTooSmallId, errs[0].GetId())
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
	// only a non-breaking change detected
	require.Equal(t, checker.RequestPropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
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
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.RequestPropertyReactivatedId, e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/test", e0.Path)
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

	config := checker.NewConfig(nil)
	changes := checker.RequestPropertyDeprecationCheck(d, osm, config)

	// Should return no changes because invalid stability causes continue
	require.Empty(t, changes)
}
