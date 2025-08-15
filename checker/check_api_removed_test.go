package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: deleting a path without deprecation is breaking
func TestBreaking_DeletedPath(t *testing.T) {
	errs := d(t, diff.NewConfig(), 1, 701)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathRemovedWithoutDeprecationId, errs[0].GetId())
	require.Equal(t, "api path removed without deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/openapi-test1.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 175, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: deleting an operation without deprecation is breaking
func TestBreaking_DeletedOp(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Put = openapi3.NewOperation()

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIRemovedWithoutDeprecationId, errs[0].GetId())
	require.Equal(t, "api removed without deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/openapi-test1.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 0, errs[0].GetBaseSource().Line)   // empty because operation was added by the code above without an origin
	require.Equal(t, 0, errs[0].GetBaseSource().Column) // empty because operation was added by the code above without an origin
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: deleting an operation after sunset date is not breaking
func TestBreaking_DeprecationPast(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-past.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deleting an operation before sunset date is breaking
func TestBreaking_RemoveBeforeSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-future.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIRemovedBeforeSunsetId, errs[0].GetId())
	require.Equal(t, "api removed before the sunset date '9999-08-10'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-future.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 11, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: deleting a deprecated operation without sunset date is not breaking
func TestBreaking_DeprecationNoSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-no-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIRemovedCheck), d, osm, checker.INFO)
	require.NoError(t, err)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIRemovedWithDeprecationId, errs[0].GetId())
	require.Equal(t, "api removed with deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, errs[0].GetLevel())
}

// BC: removing the path without a deprecation policy and without specifying sunset date is not breaking for alpha level
func TestBreaking_RemovedPathForAlpha(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)
	alpha := toJson(t, checker.STABILITY_ALPHA)
	s1.Spec.Paths.Value("/api/test").Get.Extensions["x-stability-level"] = alpha
	s1.Spec.Paths.Value("/api/test").Post.Extensions = map[string]interface{}{"x-stability-level": alpha}

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2.Spec.Paths.Delete("/api/test")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: removing the path without a deprecation policy and without specifying sunset date is not breaking for draft level
func TestBreaking_RemovedPathForDraft(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)
	draft := toJson(t, checker.STABILITY_DRAFT)
	s1.Spec.Paths.Value("/api/test").Get.Extensions["x-stability-level"] = draft
	s1.Spec.Paths.Value("/api/test").Post.Extensions = map[string]interface{}{"x-stability-level": draft}

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2.Spec.Paths.Delete("/api/test")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: removing the path without a deprecation policy and without specifying sunset date is breaking for endpoints with non draft/alpha stability level
func TestBreaking_RemovedPathForAlphaBreaking(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathRemovedWithoutDeprecationId, errs[0].GetId())
	require.Equal(t, "api path removed without deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-alpha-stability.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 7, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: removing the path without a deprecation policy and without specifying sunset date is breaking for endpoints with non draft/alpha stability level
func TestBreaking_RemovedPathForDraftBreaking(t *testing.T) {
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathRemovedWithoutDeprecationId, errs[0].GetId())
	require.Equal(t, "api path removed without deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-draft-stability.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 7, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: deleting a path after sunset date of all contained operations is not breaking
func TestBreaking_DeprecationPathPast(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-path-past.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.Empty(t, errs)
}

// BC: deleting a path with some operations having sunset date in the future is breaking
func TestBreaking_DeprecationPathMixed(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-path-mixed.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathRemovedBeforeSunsetId, errs[0].GetId())
	require.Equal(t, "api path removed before the sunset date '9999-08-10'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-path-mixed.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 18, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: deleting a path with deprecated operations without sunset date is not breaking
func TestBreaking_PathDeprecationNoSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-path-no-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.APIRemovedCheck), d, osm, checker.INFO)
	require.NoError(t, err)
	require.Len(t, errs, 2)

	require.Equal(t, checker.APIPathRemovedWithDeprecationId, errs[0].GetId())
	require.Equal(t, "api path removed with deprecation", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, errs[0].GetLevel())

	require.Equal(t, checker.APIPathRemovedWithDeprecationId, errs[1].GetId())
	require.Equal(t, "api path removed with deprecation", errs[1].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, checker.INFO, errs[1].GetLevel())
	require.Equal(t, "../data/deprecation/deprecated-path-no-sunset.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 12, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// BC: removing a deprecated enpoint with an invalid date is breaking
func TestBreaking_RemoveEndpointWithInvalidSunset(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("deprecated-invalid.yaml"))
	require.NoError(t, err)

	s2.Spec.Paths.Find("/api/test").SetOperation("GET", nil)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathSunsetParseId, errs[0].GetId())
	require.Equal(t, "failed to parse sunset date: 'sunset date doesn't conform with RFC3339: invalid-date'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-invalid.yaml", errs[0].GetSource())
	require.Equal(t, "../data/deprecation/deprecated-invalid.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 11, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}

// test sunset date without double quotes, see https://github.com/oasdiff/oasdiff/pull/198/files
func TestBreaking_DeprecationPathMixed_RFC3339_Sunset(t *testing.T) {

	s1, err := open(getDeprecationFile("deprecated-path-mixed-rfc3339-sunset.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("sunset-path.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.APIRemovedCheck), d, osm)
	require.NotEmpty(t, errs)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIPathRemovedBeforeSunsetId, errs[0].GetId())
	require.Equal(t, "api path removed before the sunset date '9999-08-10'", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/deprecated-path-mixed-rfc3339-sunset.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, 19, errs[0].GetBaseSource().Line)
	require.Equal(t, 5, errs[0].GetBaseSource().Column)
	require.Empty(t, errs[0].GetRevisionSource())
}
