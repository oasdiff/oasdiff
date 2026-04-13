package checker_test

import (
	"fmt"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------
// Helper: load stability test data files
// -------------------------------------------------------

func getStabilityFile(file string) string {
	return fmt.Sprintf("../data/stability/%s", file)
}

func stabilityConfig(sl checker.StabilityLevel, check checker.BackwardCompatibilityCheck) *checker.Config {
	config := singleCheckConfig(check)
	config.StabilityLevel = sl
	return config
}

// -------------------------------------------------------
// RequestPropertyStabilityUpdatedCheck tests
// -------------------------------------------------------

// When StabilityLevel is Beta (default), draft→stable is NOT detected because base (draft) is below threshold
func TestRequestPropertyStability_BetaLevel_DraftToStableNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelBeta, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs, "draft→stable should NOT be detected at beta threshold since base (draft) is below threshold")
}

// When StabilityLevel is Draft, detect draft→stable increase on request property
func TestRequestPropertyStability_DraftToStable_Increased(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelDraft, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyStabilityIncreasedId, errs[0].GetId())
}

// When StabilityLevel is Draft, detect stable→draft decrease on request property
func TestRequestPropertyStability_StableToDraft_Decreased(t *testing.T) {
	s1, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelDraft, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.RequestPropertyStabilityDecreasedId, errs[0].GetId())
}

// When StabilityLevel is Alpha, draft→stable is NOT reported because base (draft) is below alpha threshold
func TestRequestPropertyStability_AlphaLevel_DraftToStableNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelAlpha, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs, "draft→stable should NOT be reported when base (draft) is below alpha threshold")
}

// When no property stability actually changed, no changes emitted
func TestRequestPropertyStability_NoChange(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelDraft, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs)
}

// Nil config returns empty
func TestRequestPropertyStability_NilConfig(t *testing.T) {
	errs := checker.RequestPropertyStabilityUpdatedCheck(nil, nil, nil)
	require.Empty(t, errs)
}

// Nil diffReport.PathsDiff returns empty
func TestRequestPropertyStability_NilPathsDiff(t *testing.T) {
	config := checker.NewConfig(checker.GetAllChecks())
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.RequestPropertyStabilityUpdatedCheck(&diff.Diff{}, nil, config)
	require.Empty(t, errs)
}

// -------------------------------------------------------
// ResponsePropertyStabilityUpdatedCheck tests
// -------------------------------------------------------

// When StabilityLevel is Beta (default), draft→stable is NOT detected because base (draft) is below threshold
func TestResponsePropertyStability_BetaLevel_DraftToStableNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelBeta, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs, "draft→stable should NOT be detected at beta threshold since base (draft) is below threshold")
}

// When StabilityLevel is Draft, detect draft→stable increase on response property
func TestResponsePropertyStability_DraftToStable_Increased(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelDraft, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyStabilityIncreasedId, errs[0].GetId())
}

// When StabilityLevel is Draft, detect stable→draft decrease on response property
func TestResponsePropertyStability_StableToDraft_Decreased(t *testing.T) {
	s1, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelDraft, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.ResponsePropertyStabilityDecreasedId, errs[0].GetId())
}

// When StabilityLevel is Alpha, draft→stable is NOT reported because base (draft) is below alpha threshold
func TestResponsePropertyStability_AlphaLevel_DraftToStableNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelAlpha, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs, "draft→stable should NOT be reported when base (draft) is below alpha threshold")
}

// Nil config returns empty
func TestResponsePropertyStability_NilConfig(t *testing.T) {
	errs := checker.ResponsePropertyStabilityUpdatedCheck(nil, nil, nil)
	require.Empty(t, errs)
}

// Nil diffReport.PathsDiff returns empty
func TestResponsePropertyStability_NilPathsDiff(t *testing.T) {
	config := checker.NewConfig(checker.GetAllChecks())
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.ResponsePropertyStabilityUpdatedCheck(&diff.Diff{}, nil, config)
	require.Empty(t, errs)
}

// -------------------------------------------------------
// Endpoint-level stability change tests (checker.go)
// -------------------------------------------------------

// When StabilityLevel is Draft, detect endpoint stability decrease (stable→draft)
func TestEndpointStability_StableToDraft_Decreased(t *testing.T) {
	s1, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	requireChange(t, errs, checker.APIStabilityDecreasedId)
}

// When StabilityLevel is Draft, detect endpoint stability increase (draft→stable)
func TestEndpointStability_DraftToStable_Increased(t *testing.T) {
	s1, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	requireChange(t, errs, checker.APIStabilityIncreasedId)
}

// When StabilityLevel is Beta (default), stable→draft decrease is NOT detected because revision (draft) is below threshold
func TestEndpointStability_BetaLevel_StableToDraftNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	// StabilityLevel defaults to Beta
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	requireNoChange(t, errs, checker.APIStabilityDecreasedId)
}

// When StabilityLevel is Stable, draft→stable is NOT reported because base (draft) is below threshold
func TestEndpointStability_StableLevel_DraftToStableNotDetected(t *testing.T) {
	s1, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelStable
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	requireNoChange(t, errs, checker.APIStabilityIncreasedId)
}
