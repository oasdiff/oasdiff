package checker_test

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// CheckBackwardCompatibilityUntilLevel must NOT mutate the caller's
// diffReport. Two known mutation surfaces in the pipeline:
//   - applyStabilityLevelPolicy filters PathsDiff.Deleted /
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
