package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
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
	require.Equal(t, checker.StabilityLevelStable, checker.ParseStabilityLevel(""))
}

func TestParseStabilityLevel_Unknown(t *testing.T) {
	require.Equal(t, checker.StabilityLevelStable, checker.ParseStabilityLevel("unknown"))
}

func TestParseStabilityLevel_Ordering(t *testing.T) {
	require.True(t, checker.StabilityLevelStable > checker.StabilityLevelBeta)
	require.True(t, checker.StabilityLevelBeta > checker.StabilityLevelAlpha)
	require.True(t, checker.StabilityLevelAlpha > checker.StabilityLevelDraft)
	require.True(t, checker.StabilityLevelDraft > 0)
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

// StabilityLevelBeta (default) excludes draft and alpha (original behavior)
func TestIsIncluded_NoneThreshold_ExcludesDraftAndAlpha(t *testing.T) {
	sl := checker.StabilityLevelBeta
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
	require.Equal(t, checker.StabilityLevelBeta, checker.DefaultStabilityLevel)
}

func TestNewConfig_DefaultStabilityLevel(t *testing.T) {
	config := checker.NewConfig(checker.GetAllChecks())
	require.Equal(t, checker.StabilityLevelBeta, config.StabilityLevel)
}
