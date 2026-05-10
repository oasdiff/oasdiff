package checker_test

import (
	"slices"
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
	errs := checker.CheckBackwardCompatibility(allChecksConfig(), d, osm)
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

// CheckBackwardCompatibilityUntilLevel must NOT mutate the caller's
// diffReport. Two known mutation surfaces in the pipeline:
//   - removeDraftAndAlphaOperationsDiffs filters PathsDiff.Deleted /
//     OperationsDiff.Deleted / OperationsDiff.Modified.
//   - mergeWebhookOperationsIntoPathsDiff inserts webhook entries
//     under "webhook:..." keys into PathsDiff.Modified.
//
// Regression for #904.
func TestBreaking_CheckDoesNotMutateCallerDiff(t *testing.T) {
	s1, err := open("../data/checker/webhook_operation_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/webhook_operation_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	require.NotNil(t, d.WebhooksDiff)
	require.NotEmpty(t, d.WebhooksDiff.Modified, "fixture must have modified webhooks to exercise the merge mutation site")

	// Snapshot the parts of the diff that the check pipeline writes
	// to.
	beforeWebhooksLen := len(d.WebhooksDiff.Modified)
	var beforePathsDeleted []string
	var beforePathsModifiedLen int
	if d.PathsDiff != nil {
		beforePathsDeleted = slices.Clone(d.PathsDiff.Deleted)
		beforePathsModifiedLen = len(d.PathsDiff.Modified)
	}

	_ = checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)

	// PathsDiff.Modified must NOT carry "webhook:..." keys after the
	// call (the merge ran on a clone, not on the caller's PathsDiff).
	if d.PathsDiff != nil {
		for k := range d.PathsDiff.Modified {
			require.NotContains(t, k, "webhook:", "webhook entry leaked back into caller's PathsDiff.Modified")
		}
		require.Equal(t, beforePathsModifiedLen, len(d.PathsDiff.Modified), "PathsDiff.Modified length changed")
		require.Equal(t, beforePathsDeleted, d.PathsDiff.Deleted, "PathsDiff.Deleted slice changed")
	}
	require.Equal(t, beforeWebhooksLen, len(d.WebhooksDiff.Modified), "WebhooksDiff.Modified length changed")
}
