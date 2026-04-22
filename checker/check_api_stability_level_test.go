package checker_test

import (
	"fmt"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------
// ParseStabilityLevel tests
// -------------------------------------------------------

func TestParseStabilityLevel_Draft(t *testing.T) {
	require.Equal(t, checker.StabilityLevelDraft, checker.ParseStabilityLevel("draft"))
}

func TestParseStabilityLevel_Alpha(t *testing.T) {
	require.Equal(t, checker.StabilityLevelAlpha, checker.ParseStabilityLevel("alpha"))
}

func TestParseStabilityLevel_Beta(t *testing.T) {
	require.Equal(t, checker.StabilityLevelBeta, checker.ParseStabilityLevel("beta"))
}

func TestParseStabilityLevel_Stable(t *testing.T) {
	require.Equal(t, checker.StabilityLevelStable, checker.ParseStabilityLevel("stable"))
}

func TestParseStabilityLevel_Empty(t *testing.T) {
	require.Equal(t, checker.StabilityLevelNone, checker.ParseStabilityLevel(""))
}

func TestParseStabilityLevel_Unknown(t *testing.T) {
	require.Equal(t, checker.StabilityLevelNone, checker.ParseStabilityLevel("unknown"))
}

func TestParseStabilityLevel_Ordering(t *testing.T) {
	require.True(t, checker.StabilityLevelStable > checker.StabilityLevelBeta)
	require.True(t, checker.StabilityLevelBeta > checker.StabilityLevelAlpha)
	require.True(t, checker.StabilityLevelAlpha > checker.StabilityLevelDraft)
	require.True(t, checker.StabilityLevelDraft > checker.StabilityLevelNone)
}

// -------------------------------------------------------
// IsIncluded tests
// -------------------------------------------------------

// StabilityLevelDraft includes all levels
func TestIsIncluded_DraftThreshold_IncludesAll(t *testing.T) {
	sl := checker.StabilityLevelDraft
	require.True(t, sl.IsIncluded("draft"))
	require.True(t, sl.IsIncluded("alpha"))
	require.True(t, sl.IsIncluded("beta"))
	require.True(t, sl.IsIncluded("stable"))
	require.True(t, sl.IsIncluded("")) // empty = stable
}

// StabilityLevelAlpha excludes draft
func TestIsIncluded_AlphaThreshold_ExcludesDraft(t *testing.T) {
	sl := checker.StabilityLevelAlpha
	require.False(t, sl.IsIncluded("draft"))
	require.True(t, sl.IsIncluded("alpha"))
	require.True(t, sl.IsIncluded("beta"))
	require.True(t, sl.IsIncluded("stable"))
	require.True(t, sl.IsIncluded(""))
}

// StabilityLevelBeta excludes draft and alpha
func TestIsIncluded_BetaThreshold_ExcludesDraftAndAlpha(t *testing.T) {
	sl := checker.StabilityLevelBeta
	require.False(t, sl.IsIncluded("draft"))
	require.False(t, sl.IsIncluded("alpha"))
	require.True(t, sl.IsIncluded("beta"))
	require.True(t, sl.IsIncluded("stable"))
	require.True(t, sl.IsIncluded(""))
}

// StabilityLevelStable excludes draft, alpha, and beta
func TestIsIncluded_StableThreshold_OnlyStable(t *testing.T) {
	sl := checker.StabilityLevelStable
	require.False(t, sl.IsIncluded("draft"))
	require.False(t, sl.IsIncluded("alpha"))
	require.False(t, sl.IsIncluded("beta"))
	require.True(t, sl.IsIncluded("stable"))
	require.True(t, sl.IsIncluded(""))
}

// StabilityLevelNone (default) excludes draft and alpha (original behavior)
func TestIsIncluded_NoneThreshold_ExcludesDraftAndAlpha(t *testing.T) {
	sl := checker.StabilityLevelNone
	require.False(t, sl.IsIncluded("draft"))
	require.False(t, sl.IsIncluded("alpha"))
	require.True(t, sl.IsIncluded("beta"))
	require.True(t, sl.IsIncluded("stable"))
	require.True(t, sl.IsIncluded(""))
}

// -------------------------------------------------------
// Config defaults
// -------------------------------------------------------

func TestDefaultStabilityLevel_IsNone(t *testing.T) {
	require.Equal(t, checker.StabilityLevelNone, checker.DefaultStabilityLevel)
}

func TestNewConfig_DefaultStabilityLevel(t *testing.T) {
	config := checker.NewConfig(checker.GetAllChecks())
	require.Equal(t, checker.StabilityLevelNone, config.StabilityLevel)
}

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

// When StabilityLevel is None (default), no property stability changes are emitted
func TestRequestPropertyStability_NoneLevel_NoChanges(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelNone, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs)
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

// When StabilityLevel is Alpha, draft→stable is reported because revision meets threshold
func TestRequestPropertyStability_AlphaLevel_IncludesDraftToStable(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelAlpha, checker.RequestPropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.NotEmpty(t, errs, "draft→stable should be reported when revision meets alpha threshold")
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

// When StabilityLevel is None (default), no response property stability changes are emitted
func TestResponsePropertyStability_NoneLevel_NoChanges(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelNone, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Empty(t, errs)
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

// When StabilityLevel is Alpha, draft-level response properties are excluded
// When StabilityLevel is Alpha, draft→stable is reported because revision meets threshold
func TestResponsePropertyStability_AlphaLevel_IncludesDraftToStable(t *testing.T) {
	s1, err := open(getStabilityFile("base-property-stable-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-property-all-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := stabilityConfig(checker.StabilityLevelAlpha, checker.ResponsePropertyStabilityUpdatedCheck)
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.NotEmpty(t, errs, "draft→stable should be reported when revision meets alpha threshold")
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

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityDecreasedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected api-stability-decreased change")
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

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityIncreasedId {
			found = true
			break
		}
	}
	require.True(t, found, "expected api-stability-increased change")
}

// When StabilityLevel is None, only stable→beta decrease is detected (not stable→draft)
func TestEndpointStability_NoneLevel_OnlyStableToBeta(t *testing.T) {
	s1, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	// StabilityLevel defaults to None
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	for _, e := range errs {
		require.NotEqual(t, checker.APIStabilityDecreasedId, e.GetId(), "stable→draft should NOT be detected when StabilityLevel is None (only stable→beta)")
		require.NotEqual(t, checker.APIStabilityIncreasedId, e.GetId())
	}
}

// When StabilityLevel is Stable, draft→stable is reported because revision meets threshold
func TestEndpointStability_StableLevel_IncludesDraftToStable(t *testing.T) {
	s1, err := open(getStabilityFile("revision-endpoint-draft.yaml"))
	require.NoError(t, err)
	s2, err := open(getStabilityFile("base-endpoint-stable.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelStable
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityIncreasedId {
			found = true
		}
	}
	require.True(t, found, "draft→stable should be reported when revision meets stable threshold")
}
