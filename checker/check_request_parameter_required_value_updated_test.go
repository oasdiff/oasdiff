package checker_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// BC: changing an existing header param from optional to required is breaking
func TestBreaking_HeaderParamBecameRequired(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRequiredValueUpdatedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterBecomeRequiredId,
		Args:      []any{"header", "network-policies"},
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/api/{domain}/{project}/install-command",
		Source:    load.NewSource("../data/openapi-test1.yaml"),
	}, errs[0])
}

// BC: changing an existing header param from optional to required with source tracking
func TestBreaking_HeaderParamBecameRequired_WithSources(t *testing.T) {
	enableOriginTracking(t)
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(singleCheckConfig(checker.RequestParameterRequiredValueUpdatedCheck), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestParameterBecomeRequiredId, errs[0].GetId())

	// Verify source tracking for parameter-level checkers
	require.NotEmpty(t, errs[0].GetBaseSource())
	require.NotEmpty(t, errs[0].GetRevisionSource())
	require.Equal(t, "../data/openapi-test1.yaml", errs[0].GetBaseSource().File)
	require.Equal(t, "../data/openapi-test1.yaml", errs[0].GetRevisionSource().File)
	require.NotZero(t, errs[0].GetBaseSource().Line)
	require.NotZero(t, errs[0].GetRevisionSource().Line)
}

// CL: changing an existing header param from required to optional
func TestBreaking_HeaderParamBecameOptional(t *testing.T) {
	s1 := l(t, 1)
	s2 := l(t, 1)

	s1.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = true
	s2.Spec.Paths.Value(installCommandPath).Get.Parameters.GetByInAndName(openapi3.ParameterInHeader, "network-policies").Required = false

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.RequestParameterRequiredValueUpdatedCheck), d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ApiChange{
		Id:        checker.RequestParameterBecomeOptionalId,
		Args:      []any{"header", "network-policies"},
		Level:     checker.INFO,
		Operation: "GET",
		Path:      "/api/{domain}/{project}/install-command",
		Source:    load.NewSource("../data/openapi-test1.yaml"),
	}, errs[0])
}
