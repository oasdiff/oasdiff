package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// decreasing stability level is breaking
func TestBreaking_StabilityLevelDecreased(t *testing.T) {

	s1, err := open(deprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("base-alpha-stability.yaml"))
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

// increasing stability level is not breaking
func TestBreaking_StabilityLevelIncreased(t *testing.T) {

	s1, err := open(deprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Empty(t, errs)
}

// specifying an invalid stability level in revision is breaking
func TestBreaking_InvalidStabilityLevelInRevision(t *testing.T) {
	s1, err := open(deprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", requireChange(t, errs, checker.APIInvalidStabilityLevelId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability.yaml", errs[0].GetSource())
}

// specifying an invalid stability level in base is breaking
func TestBreaking_InvalidStabilityLevelInBase(t *testing.T) {
	s1, err := open(deprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", requireChange(t, errs, checker.APIInvalidStabilityLevelId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability.yaml", errs[0].GetSource())
}

// specifying a non-text, not-json stability level in base is breaking
func TestBreaking_InvalidNonJsonStabilityLevel(t *testing.T) {
	s1, err := open(deprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(deprecationFile("base-invalid-stability-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	require.Len(t, errs, 1)
	require.Equal(t, "failed to parse stability level: `x-stability-level isn't a string nor valid json`", requireChange(t, errs, checker.APIInvalidStabilityLevelId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "../data/deprecation/base-invalid-stability-2.yaml", errs[0].GetSource())
}

// -------------------------------------------------------
// Endpoint-level stability change tests
// -------------------------------------------------------

// At Draft, an endpoint stability decrease (stable→draft) is reported.
func TestEndpointStability_StableToDraft_Decreased(t *testing.T) {
	errs := stabilityChanges(t, "base-endpoint-stable.yaml", "revision-endpoint-draft.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.APIStabilityDecreasedId)
}

// At Draft, an endpoint stability increase (draft→stable) is reported.
func TestEndpointStability_DraftToStable_Increased(t *testing.T) {
	errs := stabilityChanges(t, "revision-endpoint-draft.yaml", "base-endpoint-stable.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.APIStabilityIncreasedId)
}

// At Beta (default), a stable→draft decrease IS reported: the base (stable) is
// within the threshold, so leaving it is reported. Gating on the lower (draft)
// level instead would drop this and regress the existing api-stability-decreased ERR.
func TestEndpointStability_BetaLevel_StableToDraftDetected(t *testing.T) {
	errs := stabilityChanges(t, "base-endpoint-stable.yaml", "revision-endpoint-draft.yaml", checker.StabilityLevelBeta)
	requireChange(t, errs, checker.APIStabilityDecreasedId)
}

// At Stable, draft→stable is not reported because base (draft) is below the threshold.
func TestEndpointStability_StableLevel_DraftToStableNotDetected(t *testing.T) {
	errs := stabilityChanges(t, "revision-endpoint-draft.yaml", "base-endpoint-stable.yaml", checker.StabilityLevelStable)
	requireNoChange(t, errs, checker.APIStabilityIncreasedId)
}
