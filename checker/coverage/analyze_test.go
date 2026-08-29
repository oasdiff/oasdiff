package coverage_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/stretchr/testify/require"
)

func find(t *testing.T, edits []coverage.Edit, location, action string) coverage.Edit {
	t.Helper()
	for _, edit := range edits {
		if edit.Location == location && edit.Action == action {
			return edit
		}
	}
	t.Fatalf("no edit %s %s", location, action)
	return coverage.Edit{}
}

// a rule covers the edits its claim names, and only those
func TestAnalyze_Covered(t *testing.T) {
	edits := coverage.Analyze([]rules.Rule{{
		Id:        "test-rule",
		Locations: []string{"paths.*.*.requestBody.content.*.schema.maxLength:decrease"},
	}})

	covered := find(t, edits, "paths.*.*.requestBody.content.*.schema.maxLength", "decrease")
	require.Equal(t, coverage.Covered, covered.Status)
	require.Equal(t, []string{"test-rule"}, covered.Checks)
	require.Empty(t, covered.SuggestedId, "a covered edit needs no suggestion")
	require.Empty(t, covered.Reason)

	sibling := find(t, edits, "paths.*.*.requestBody.content.*.schema.maxLength", "increase")
	require.NotEqual(t, coverage.Covered, sibling.Status, "the claim names one action")
}

// an edit no rule claims and no waiver explains fails the audit, and carries
// a suggested id for the check that would cover it. Removing a path is
// checked in practice, so no waiver accounts for it: passing no rules is
// what leaves it uncovered here.
func TestAnalyze_Uncovered(t *testing.T) {
	edits := coverage.Analyze(nil)

	uncovered := find(t, edits, "paths.*", "remove")
	require.Equal(t, coverage.Uncovered, uncovered.Status)
	require.Empty(t, uncovered.Checks)
	require.Equal(t, "api-path-removed", uncovered.SuggestedId)
}

// a waiver explains an unchecked edit; only an open waiver implies a check
// that could exist. Both edits below are waived whatever rules are passed,
// since a waiver exists exactly where the checks do not.
func TestAnalyze_Waived(t *testing.T) {
	edits := coverage.Analyze(nil)

	open := find(t, edits, "webhooks.*.*.requestBody.content.*.schema.maxLength", "decrease")
	require.Equal(t, coverage.Waived, open.Status)
	require.Equal(t, coverage.CategoryOpen, open.Category)
	require.NotEmpty(t, open.Reason)
	require.Equal(t, "webhook-max-length-decreased", open.SuggestedId)

	resolved := find(t, edits, "components.schemas.*.maxLength", "decrease")
	require.Equal(t, coverage.Waived, resolved.Status)
	require.Equal(t, coverage.CategoryResolvedAtUsage, resolved.Category)
	require.Empty(t, resolved.SuggestedId, "the check belongs at the usage sites, which have their own edits")
}

// an edit that cannot change which payloads are valid expects no check
func TestAnalyze_NonContract(t *testing.T) {
	edits := coverage.Analyze(nil)

	for _, tc := range []struct{ location, action, reason string }{
		{"paths.*.*.description", "change", "annotation: documentation-only field"},
		{"paths.*.*.x-*", "change", "specification extension"},
		{"paths.*.*.servers.*.url", "change", "server URLs are deployment metadata, not part of the request/response contract"},
	} {
		edit := find(t, edits, tc.location, tc.action)
		require.Equal(t, coverage.NonContract, edit.Status, tc.location)
		require.Equal(t, tc.reason, edit.Reason, tc.location)
		require.Empty(t, edit.SuggestedId, tc.location)
	}
}

// every edit is decided, and the polarity travels with it
func TestAnalyze_EveryEditDecided(t *testing.T) {
	edits := coverage.Analyze(nil)
	require.NotEmpty(t, edits)

	for _, edit := range edits {
		require.Contains(t,
			[]coverage.Status{coverage.Covered, coverage.Uncovered, coverage.Waived, coverage.NonContract},
			edit.Status, edit.Location)
		require.NotEmpty(t, edit.Polarity, edit.Location)
	}
}
