package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------
// RequestPropertyStabilityUpdatedCheck tests
// -------------------------------------------------------

// At Beta (default), draft→stable is not reported because base (draft) is below the threshold.
func TestRequestPropertyStability_BetaLevel_DraftToStableNotDetected(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelBeta)
	requireNoChange(t, errs, checker.RequestPropertyStabilityIncreasedId)
}

// At Draft, a draft→stable increase on a request property is reported.
func TestRequestPropertyStability_DraftToStable_Increased(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.RequestPropertyStabilityIncreasedId)
}

// At Draft, a stable→draft decrease on a request property is reported.
func TestRequestPropertyStability_StableToDraft_Decreased(t *testing.T) {
	errs := stabilityChanges(t, "revision-property-all-stable.yaml", "base-property-stable-draft.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.RequestPropertyStabilityDecreasedId)
}

// At Alpha, draft→stable is not reported because base (draft) is below the threshold.
func TestRequestPropertyStability_AlphaLevel_DraftToStableNotDetected(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelAlpha)
	requireNoChange(t, errs, checker.RequestPropertyStabilityIncreasedId)
}

// When no property stability actually changed, no stability change is reported.
func TestRequestPropertyStability_NoChange(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "base-property-stable-draft.yaml", checker.StabilityLevelDraft)
	requireNoChange(t, errs, checker.RequestPropertyStabilityIncreasedId)
	requireNoChange(t, errs, checker.RequestPropertyStabilityDecreasedId)
}

// An unparseable x-stability-level on a request property is reported as invalid,
// since checkInvalidStabilityLevels does not descend into property schemas.
func TestRequestPropertyStability_InvalidLevel(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-invalid.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.APIInvalidStabilityLevelId)
}

// A property change on an endpoint below the threshold is not reported: the
// operation is out of scope, so its property changes are filtered out with it.
func TestRequestPropertyStability_BelowThresholdEndpoint_NotReported(t *testing.T) {
	errs := stabilityChanges(t, "base-draft-endpoint-property.yaml", "revision-draft-endpoint-property.yaml", checker.StabilityLevelBeta)
	requireNoChange(t, errs, checker.RequestPropertyStabilityDecreasedId)
	requireNoChange(t, errs, checker.RequestPropertyStabilityIncreasedId)
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

// At Beta (default), draft→stable is not reported because base (draft) is below the threshold.
func TestResponsePropertyStability_BetaLevel_DraftToStableNotDetected(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelBeta)
	requireNoChange(t, errs, checker.ResponsePropertyStabilityIncreasedId)
}

// At Draft, a draft→stable increase on a response property is reported.
func TestResponsePropertyStability_DraftToStable_Increased(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.ResponsePropertyStabilityIncreasedId)
}

// At Draft, a stable→draft decrease on a response property is reported.
func TestResponsePropertyStability_StableToDraft_Decreased(t *testing.T) {
	errs := stabilityChanges(t, "revision-property-all-stable.yaml", "base-property-stable-draft.yaml", checker.StabilityLevelDraft)
	requireChange(t, errs, checker.ResponsePropertyStabilityDecreasedId)
}

// At Alpha, draft→stable is not reported because base (draft) is below the threshold.
func TestResponsePropertyStability_AlphaLevel_DraftToStableNotDetected(t *testing.T) {
	errs := stabilityChanges(t, "base-property-stable-draft.yaml", "revision-property-all-stable.yaml", checker.StabilityLevelAlpha)
	requireNoChange(t, errs, checker.ResponsePropertyStabilityIncreasedId)
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
