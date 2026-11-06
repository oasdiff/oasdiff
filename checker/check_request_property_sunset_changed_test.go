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

// deleting the sunset date of a deprecated request property is breaking; the change points at the x-sunset field in the base spec
func TestRequestPropertySunsetChanged_Deleted(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_no_sunset.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertySunsetChangedCheck), d, osm)
	change := requireSingleChange(t, errs, checker.RequestPropertySunsetDeletedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "request property `oldField` sunset date deleted, but deprecated=true kept", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, deprecationFile("property_deprecated_future.yaml"), change.GetBaseSource().File)
	require.Equal(t, 36, change.GetBaseSource().Line)
	require.Empty(t, change.GetRevisionSource())
}

// changing another field of a deprecated request property is not reported
func TestRequestPropertySunsetChanged_SunsetUnchanged(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertySunsetChangedCheck, deprecationFile("property_deprecated_future.yaml"), deprecationFile("property_deprecated.yaml")))
}

// deprecating a request property while deleting its sunset date is not reported here: the deprecation check reports it
func TestRequestPropertySunsetChanged_DeletedWhileDeprecating(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertySunsetChangedCheck, deprecationFile("property_sunset_not_deprecated.yaml"), deprecationFile("property_deprecated_no_sunset.yaml")))
}

// deleting another extension of a deprecated request property without a sunset date is not reported
func TestRequestPropertySunsetChanged_OtherExtensionDeleted(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertySunsetChangedCheck, deprecationFile("property_deprecated_no_sunset_other_extension.yaml"), deprecationFile("property_deprecated_no_sunset.yaml")))
}

// changing the sunset date of a deprecated request property to an earlier date than the deprecation policy allows is breaking
func TestRequestPropertySunsetChanged_TooSmall(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)
	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertySunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm)
	change := requireSingleChange(t, errs, checker.RequestPropertySunsetDateChangedTooSmallId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, fmt.Sprintf("request property `oldField` sunset date changed to an earlier date, from `2099-12-31` to `%s`, new sunset date must be not earlier than `2099-12-31` and at least `10` days from now", sunsetDate), change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing the sunset date of a deprecated request property to an invalid date is breaking; the change points at the x-sunset field in each spec
func TestRequestPropertySunsetChanged_ToInvalid(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_invalid_sunset.yaml"), newLoaderWithOriginTracking())
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestPropertySunsetChangedCheck), d, osm)
	change := requireSingleChange(t, errs, checker.RequestPropertyDeprecatedInvalidId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "request property `oldField` deprecated with invalid sunset date: `sunset date doesn't conform with RFC3339: invalid-date-format`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, deprecationFile("property_deprecated_future.yaml"), change.GetBaseSource().File)
	require.Equal(t, 36, change.GetBaseSource().Line)
	require.Equal(t, deprecationFile("property_deprecated_invalid_sunset.yaml"), change.GetRevisionSource().File)
	require.Equal(t, 36, change.GetRevisionSource().Line)
}

// changing an invalid sunset date of a deprecated request property to a valid
// one is not reported, even to one sooner than the deprecation policy allows
func TestRequestPropertySunsetChanged_FromInvalid(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_invalid_sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_invalid_sunset.yaml"))
	require.NoError(t, err)
	s2.Spec.Components.Schemas["TestRequest"].Value.Properties["oldField"].Value.Extensions[diff.SunsetExtension] = toJson(t, civil.DateOf(time.Now()).AddDays(9).String())

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestPropertySunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm, checker.INFO))
}

// a sunset date change under an invalid stability level is skipped (handled in CheckBackwardCompatibility)
func TestRequestPropertySunsetChanged_WithInvalidStability(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_no_sunset.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/test").Post.Extensions[diff.XStabilityLevelExtension] = toJson(t, "invalid-stability")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, checker.RequestPropertySunsetChangedCheck(d, osm, checker.NewConfig(nil)))
}
