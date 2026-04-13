package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// BC: decreasing stability level is breaking
func TestBreaking_StabilityLevelDecreased(t *testing.T) {

	s1, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// Use Alpha threshold so both beta and alpha are at/above threshold
	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelAlpha
	errs := checker.CheckBackwardCompatibility(config, d, osm)
	require.Len(t, errs, 1)

	require.IsType(t, checker.ApiChange{}, errs[0])
	e0 := errs[0].(checker.ApiChange)
	require.Equal(t, checker.APIStabilityDecreasedId, e0.Id)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Equal(t, "endpoint stability level decreased from `beta` to `alpha`", e0.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// BC: increasing stability level is not breaking
func TestBreaking_StabilityLevelIncreased(t *testing.T) {

	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// BC: specifying an invalid stability level in revision is breaking
func TestBreaking_InvalidStabilityLevelInRevision(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability.yaml", errs[0].GetSource())
}

// BC: specifying an invalid stability level in base is breaking
func TestBreaking_InvalidStabilityLevelInBase(t *testing.T) {
	s1, err := open(getDeprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability.yaml", errs[0].GetSource())
}

// BC: specifying a non-text, not-json stability level in base is breaking
func TestBreaking_InvalidNonJsonStabilityLevel(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-invalid-stability-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `x-stability-level isn't a string nor valid json`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability-2.yaml", errs[0].GetSource())
}
