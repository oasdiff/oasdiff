package checker_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// detecting deprecated response properties with sunset date
func TestResponsePropertyDeprecationCheck(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.ResponsePropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.ResponsePropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated response property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected ResponsePropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// detecting deprecated response properties in allOf schemas
func TestResponsePropertyDeprecationCheck_AllOf(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.ResponsePropertyDeprecationCheck(d, osm, config)

	found := false
	for _, c := range changes {
		if c.GetId() == checker.ResponsePropertyDeprecatedId {
			found = true
			t.Logf("Found deprecated response allOf property: %+v", c)
		}
	}
	if !found {
		t.Errorf("Expected ResponsePropertyDeprecatedId in changes, got: %+v", changes)
	}
}

// Ensuring no duplicate deprecation reports for the same response property
func TestResponsePropertyDeprecationCheck_NoDuplicates(t *testing.T) {
	s1, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_base.yaml"))
	require.NoError(t, err)
	s2, err := open(getPropertyDeprecationFile("response_property_deprecation_allof_spec.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := checker.NewConfig(nil)
	changes := checker.ResponsePropertyDeprecationCheck(d, osm, config)

	// Count occurrences of each property
	propCount := make(map[string]int)
	for _, c := range changes {
		if c.GetId() == checker.ResponsePropertyDeprecatedId {
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

	// Update the sunset date in the spec to be too small
	s2.Spec.Components.Schemas["TestResponse"].Value.Properties["legacyField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	c := singleCheckConfig(checker.ResponsePropertyDeprecationCheck).WithDeprecation(0, 10)
	errs := checker.CheckBackwardCompatibility(c, d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertySunsetDateTooSmallId, errs[0].GetId())
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
	// only a non-breaking change detected
	require.Equal(t, checker.ResponsePropertyDeprecatedId, errs[0].GetId())
	require.Equal(t, checker.INFO, errs[0].GetLevel())
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
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.ResponsePropertyReactivatedId, e0.Id)
	require.Equal(t, "POST", e0.Operation)
	require.Equal(t, "/test", e0.Path)
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
	require.NotEmpty(t, errs)
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

	config := checker.NewConfig(nil)
	changes := checker.ResponsePropertyDeprecationCheck(d, osm, config)

	// Should return no changes because invalid stability causes continue
	require.Empty(t, changes)
}
