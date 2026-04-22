package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------
// Endpoint stability level decrease/increase
// -------------------------------------------------------

// Decreasing stability level (beta→alpha) is detected when StabilityLevel is set to draft
func TestBreaking_StabilityLevelDecreased(t *testing.T) {
	s1, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	var found checker.Change
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityDecreasedId {
			found = e
			break
		}
	}
	require.NotNil(t, found, "expected api-stability-decreased change")
	require.IsType(t, checker.ApiChange{}, found)
	e0 := found.(checker.ApiChange)
	require.Equal(t, "GET", e0.Operation)
	require.Equal(t, "/api/test", e0.Path)
	require.Equal(t, "endpoint stability level decreased from `beta` to `alpha`", e0.GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// Decreasing stability level (stable→draft) detected when both explicitly set
func TestBreaking_StabilityLevelDecreased_BetaToDraft(t *testing.T) {
	s1, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-draft-stability.yaml"))
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
	require.True(t, found, "expected api-stability-decreased change for beta→draft")
}

// Increasing stability level (alpha→beta) detected
func TestBreaking_StabilityLevelIncreased_AlphaToBeta(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
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
	require.True(t, found, "expected api-stability-increased change for alpha→beta")
}

// Increasing stability level (draft→beta) detected
func TestBreaking_StabilityLevelIncreased_DraftToBeta(t *testing.T) {
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
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
	require.True(t, found, "expected api-stability-increased change for draft→beta")
}

// -------------------------------------------------------
// StabilityLevel=None means NO stability detection
// -------------------------------------------------------

// When StabilityLevel is None (default), stability decrease is NOT detected
func TestBreaking_StabilityLevelNone_OnlyStableToBetaDetected(t *testing.T) {
	// beta→alpha should NOT be detected without flag
	s1, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	for _, e := range errs {
		require.NotEqual(t, checker.APIStabilityDecreasedId, e.GetId(), "beta→alpha should not be detected when StabilityLevel is None")
		require.NotEqual(t, checker.APIStabilityIncreasedId, e.GetId(), "stability increase should not be detected when StabilityLevel is None")
	}
}

// When StabilityLevel is None, increasing stability is not detected either
func TestBreaking_StabilityLevelNone_NoIncreaseDetected(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
	for _, e := range errs {
		require.NotEqual(t, checker.APIStabilityIncreasedId, e.GetId())
	}
}

// -------------------------------------------------------
// Threshold filtering: report if either base or revision meets the threshold
// -------------------------------------------------------

// When StabilityLevel is Stable, draft→stable IS reported (revision meets threshold)
func TestBreaking_StabilityLevelStable_IncludesDraftToStable(t *testing.T) {
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base.yaml"))
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

// When StabilityLevel is Alpha, draft→alpha IS reported (revision meets threshold)
func TestBreaking_StabilityLevelAlpha_IncludesDraftToAlpha(t *testing.T) {
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelAlpha
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityIncreasedId {
			found = true
		}
	}
	require.True(t, found, "draft→alpha should be reported when revision meets alpha threshold")
}

// When StabilityLevel is Beta, alpha→beta IS reported (revision meets threshold)
func TestBreaking_StabilityLevelBeta_IncludesAlphaToBeta(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelBeta
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityIncreasedId {
			found = true
		}
	}
	require.True(t, found, "alpha→beta should be reported when revision meets beta threshold")
}

// When StabilityLevel is Alpha, alpha→beta change IS reported (alpha is within alpha threshold)
func TestBreaking_StabilityLevelAlpha_IncludesAlphaToBeta(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-beta-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelAlpha
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	found := false
	for _, e := range errs {
		if e.GetId() == checker.APIStabilityIncreasedId {
			found = true
			break
		}
	}
	require.True(t, found, "alpha→beta should be reported when threshold is alpha")
}

// -------------------------------------------------------
// Invalid stability level tests
// -------------------------------------------------------

// Specifying an invalid stability level in revision is detected when StabilityLevel is set
func TestBreaking_InvalidStabilityLevelInRevision(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// Specifying an invalid stability level in base is detected when StabilityLevel is set
func TestBreaking_InvalidStabilityLevelInBase(t *testing.T) {
	s1, err := open(getDeprecationFile("base-invalid-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `value is not one of draft, alpha, beta or stable: \"invalid\"`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// Specifying a non-text, not-json stability level is detected when StabilityLevel is set
func TestBreaking_InvalidNonJsonStabilityLevel(t *testing.T) {
	s1, err := open(getDeprecationFile("base.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-invalid-stability-2.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	require.Len(t, errs, 1)
	require.Equal(t, checker.APIInvalidStabilityLevelId, errs[0].GetId())
	require.Equal(t, "failed to parse stability level: `x-stability-level isn't a string nor valid json`", errs[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

// -------------------------------------------------------
// Same stability level = no change
// -------------------------------------------------------

// Same stability (alpha→alpha) produces no stability change
func TestBreaking_SameStabilityLevel_NoChange(t *testing.T) {
	s1, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	s2, err := open(getDeprecationFile("base-alpha-stability.yaml"))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	for _, e := range errs {
		require.NotEqual(t, checker.APIStabilityDecreasedId, e.GetId())
		require.NotEqual(t, checker.APIStabilityIncreasedId, e.GetId())
	}
}

// -------------------------------------------------------
// Draft/alpha deletion filtering
// -------------------------------------------------------

// Deleting a path where ALL operations are draft is filtered at Stable threshold
func TestBreaking_DeleteDraftEndpoint_StableThreshold_NotReported(t *testing.T) {
	// Create a spec with only a single draft GET operation
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)
	// Remove POST so only draft GET remains
	s1.Spec.Paths.Value("/api/test").Post = nil

	s2, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Delete("/api/test")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelStable
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	for _, e := range errs {
		require.NotContains(t, e.GetId(), "path-removed", "draft-only path deletion should be filtered out at stable threshold")
	}
}

// Deleting a path where ALL operations are draft IS kept in diff at Draft threshold
func TestBreaking_DeleteDraftEndpoint_DraftThreshold_Reported(t *testing.T) {
	s1, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)
	// Remove POST so only draft GET remains
	s1.Spec.Paths.Value("/api/test").Post = nil

	s2, err := open(getDeprecationFile("base-draft-stability.yaml"))
	require.NoError(t, err)
	s2.Spec.Paths.Delete("/api/test")

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	// Verify the path is in the deleted list before filtering
	require.NotEmpty(t, d.PathsDiff.Deleted, "path should be in deleted list")

	config := allChecksConfig()
	config.StabilityLevel = checker.StabilityLevelDraft
	// Run the checker — this calls removeDraftAndAlphaOperationsDiffs which filters deleted paths
	_ = checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)

	// After filtering with Draft threshold, the draft path should still be in the deleted list
	require.NotEmpty(t, d.PathsDiff.Deleted, "draft path should NOT be filtered out at draft threshold")
}
