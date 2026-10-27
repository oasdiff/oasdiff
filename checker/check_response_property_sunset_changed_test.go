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

// deleting the sunset date of a deprecated response property is breaking
func TestResponsePropertySunsetChanged_Deleted(t *testing.T) {
	errs := checkChanges(t, checker.ResponsePropertySunsetChangedCheck, deprecationFile("property_deprecated_future.yaml"), deprecationFile("property_deprecated_no_sunset.yaml"))
	change := requireSingleChange(t, errs, checker.ResponsePropertySunsetDeletedId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "response property `legacyField` sunset date deleted, but deprecated=true kept", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing the sunset date of a deprecated response property to an earlier date than the deprecation policy allows is breaking
func TestResponsePropertySunsetChanged_TooSmall(t *testing.T) {
	s1, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("property_deprecated_future.yaml"))
	require.NoError(t, err)
	sunsetDate := civil.DateOf(time.Now()).AddDays(9).String()
	s2.Spec.Components.Schemas["TestResponse"].Value.Properties["legacyField"].Value.Extensions[diff.SunsetExtension] = toJson(t, sunsetDate)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.ResponsePropertySunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm)
	change := requireSingleChange(t, errs, checker.ResponsePropertySunsetDateChangedTooSmallId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, fmt.Sprintf("response property `legacyField` sunset date changed to an earlier date, from `2099-12-31` to `%s`, new sunset date must be not earlier than `2099-12-31` and at least `10` days from now", sunsetDate), change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing the sunset date of a deprecated response property to an invalid date is breaking
func TestResponsePropertySunsetChanged_ToInvalid(t *testing.T) {
	errs := checkChanges(t, checker.ResponsePropertySunsetChangedCheck, deprecationFile("property_deprecated_future.yaml"), deprecationFile("property_deprecated_invalid_sunset.yaml"))
	change := requireSingleChange(t, errs, checker.ResponsePropertyDeprecatedInvalidId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, "response property `legacyField` deprecated with invalid sunset date: `sunset date doesn't conform with RFC3339: invalid-date-format`", change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}
