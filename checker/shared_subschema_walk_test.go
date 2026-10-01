package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// A change to one referenced component is one contract change for an operation,
// regardless of how many properties provide paths to that component.
func TestResponseSharedSchemaChangeReportedOncePerOperation(t *testing.T) {
	base, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, checker.ResponseOptionalPropertyAddedId, changes[0].GetId())
}

func TestResponseSharedSchemaUsesStableRepresentativePath(t *testing.T) {
	base, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	for range 30 {
		d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
		require.NoError(t, err)
		changes := checker.CheckBackwardCompatibilityUntilLevel(
			singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
			d, sources, checker.INFO,
		)
		require.Len(t, changes, 1)
		change, ok := changes[0].(checker.ApiChange)
		require.True(t, ok)
		require.Equal(t, "left/extra", change.Args[0])
	}
}

// Root reaches Shared through both left and right, so a change to Shared
// alters the contract at two places in the payload, and each is reported with
// its own path.
func TestResponseSharedSchemaReportsEveryPath(t *testing.T) {
	base, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)

	paths := []string{}
	for _, change := range changes {
		apiChange, ok := change.(checker.ApiChange)
		require.True(t, ok)
		paths = append(paths, apiChange.Args[0].(string))
	}
	require.ElementsMatch(t, []string{"left/extra", "right/extra"}, paths)
}
