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

// deleting sunset header for a deprecated endpoint is breaking
func TestBreaking_SunsetDeletedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(deprecationFile("deprecated-with-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "API sunset date deleted, but deprecated=true kept", requireChange(t, errs, checker.APISunsetDeletedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// the sunset change points at the x-sunset field in the base spec
func TestBreaking_SunsetDeleted_WithSources(t *testing.T) {
	errs := checkChangesWithOrigins(t, checker.APISunsetChangedCheck, deprecationFile("deprecated-with-sunset.yaml"), deprecationFile("deprecated-no-sunset.yaml"))
	require.Len(t, errs, 1)

	require.Equal(t, checker.APISunsetDeletedId, errs[0].GetId())
	require.NotEmpty(t, errs[0].GetBaseSource())
	require.Equal(t, deprecationFile("deprecated-with-sunset.yaml"), errs[0].GetBaseSource().File)
	require.Equal(t, 13, errs[0].GetBaseSource().Line) // the x-sunset field
	require.Empty(t, errs[0].GetRevisionSource())
}

// changing sunset to an earlier date for a deprecated endpoint with a deprecation policy is breaking
func TestBreaking_SunsetModifiedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck, checker.WithDeprecation(31, 180)), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "API sunset date changed to an earlier date, from `9999-08-10` to `2022-08-10`, new sunset date must be not earlier than `9999-08-10` and at least `180` days from now", requireChange(t, errs, checker.APISunsetDateChangedTooSmallId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing sunset to an invalid date for a deprecated endpoint is breaking
func TestBreaking_SunsetModifiedToInvalidForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, "sunset date doesn't conform with RFC3339: `\"invalid-date\"`", requireChange(t, errs, checker.APIPathSunsetParseId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// changing sunset from an invalid date for a deprecated endpoint is breaking
// changing an invalid sunset date of a deprecated endpoint to a valid one is
// not reported, as clients had no date to rely on
func TestBreaking_SunsetModifiedFromInvalidForDeprecatedEndpoint(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.APISunsetChangedCheck, deprecationFile("deprecated-invalid.yaml"), deprecationFile("deprecated-future.yaml")))
}

// deleting other extension (not sunset) header for a deprecated endpoint is not breaking
func TestBreaking_NonSunsetDeletedForDeprecatedEndpoint(t *testing.T) {

	s1, err := open(deprecationFile("deprecated-with-other-extension.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// no change to headers for a deprecated endpoint is not breaking
func TestBreaking_NoChangeToSunsetDeprecatedEndpoint(t *testing.T) {

	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-future-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)
	require.Empty(t, errs)
}

// a sunset date that fails to parse on one side only is reported against that
// side, so the side that parsed carries no source.
func TestSunsetChanged_ParseFailureIsOneSided(t *testing.T) {
	loader := newLoaderWithOriginTracking()

	s1, err := open(deprecationFile("deprecated-future.yaml"), loader)
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-invalid.yaml"), loader)
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck), d, osm)

	change := requireChange(t, errs, checker.APIPathSunsetParseId)
	require.Nil(t, change.GetBaseSource(), "the base's sunset date parsed, so it is not part of this finding")
	require.NotNil(t, change.GetRevisionSource())
	require.Equal(t, 13, change.GetRevisionSource().Line) // the revision's x-sunset, which is the unparseable one
}

// deleting the sunset date of a deprecated endpoint is reported even when the
// stability level does not parse, which only the grace period needs
func TestSunsetChanged_DeletedWithInvalidStability(t *testing.T) {
	s1, err := open(deprecationFile("deprecated-with-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/test").Get.Extensions = map[string]any{diff.XStabilityLevelExtension: toJson(t, "invalid-stability")}

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.APISunsetChangedCheck(d, osm, checker.NewConfig(nil))
	requireSingleChange(t, errs, checker.APISunsetDeletedId)
}

// moving a sunset date earlier is breaking even when the new date still
// respects the deprecation policy: clients scheduled against the old one
func TestSunsetChanged_Earlier(t *testing.T) {
	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	earlier := civil.DateOf(time.Now()).AddDays(100).String()
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, earlier)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm)
	change := requireSingleChange(t, errs, checker.APISunsetDateChangedEarlierId)
	require.Equal(t, checker.ERR, change.GetLevel())
	require.Equal(t, fmt.Sprintf("API sunset date changed to an earlier date, from `9999-08-10` to `%s`", earlier), change.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// a sunset date moved earlier is reported even when the stability level does
// not parse, since only the too-small comparison needs the grace period
func TestSunsetChanged_EarlierWithInvalidStability(t *testing.T) {
	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, civil.DateOf(time.Now()).AddDays(100).String())
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.XStabilityLevelExtension] = toJson(t, "invalid-stability")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.APISunsetChangedCheck(d, osm, checker.NewConfig(nil))
	requireSingleChange(t, errs, checker.APISunsetDateChangedEarlierId)
}

// moving a sunset date later is not reported
func TestSunsetChanged_Later(t *testing.T) {
	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, "9999-12-31")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.Empty(t, checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APISunsetChangedCheck), d, osm, checker.INFO))
}

// the earlier-date finding carries the reason, since a provider with no
// deprecation policy would not expect it
func TestSunsetChanged_EarlierCarriesTheReason(t *testing.T) {
	s1, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Value("/api/test").Get.Extensions[diff.SunsetExtension] = toJson(t, civil.DateOf(time.Now()).AddDays(100).String())

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APISunsetChangedCheck, checker.WithDeprecation(0, 10)), d, osm)
	change := requireSingleChange(t, errs, checker.APISunsetDateChangedEarlierId)
	require.Contains(t, change.GetComment(checker.NewDefaultLocalizer()), "treats a published x-sunset date as a commitment")
}
