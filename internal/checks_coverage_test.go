package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/checker/coverage"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func TestChecksCoverage_TextFormats(t *testing.T) {
	for _, format := range []string{"text", "json", "yaml"} {
		var stdout bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format "+format), &stdout, &stdout), format)
		require.NotEmpty(t, stdout.String(), format)
	}
}

// the audit passes, so the uncovered listing is empty and renders as a
// valid empty document
func TestChecksCoverage_UncoveredIsEmpty(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --tags uncovered"), &stdout, &stdout))
	require.JSONEq(t, "[]", stdout.String())
}

func TestChecksCoverage_TagsFilter(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --tags covered,request,decrease"), &stdout, &stdout))

	var rows []coverage.Edit
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	for _, row := range rows {
		require.Equal(t, coverage.Covered, row.Status)
		require.Equal(t, "request", row.Polarity)
		require.Equal(t, "decrease", row.Action)
		require.NotEmpty(t, row.Checks)
	}
}

func TestChecksCoverage_Patterns(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json --patterns"), &stdout, &stdout))

	var rows []coverage.Pattern
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	for _, row := range rows {
		require.Contains(t, []string{"waiver", "non-contract"}, row.Kind)
		require.Positive(t, row.Edits, "stale pattern %q accounts for no edits", row.Pattern)
	}
}

// Same vocabulary-driven sweep for the coverage listing, against a single
// run: coverage tags match by equality on the status, polarity, and action
// fields, so the rows decide each tag without rerunning the command. The
// uncovered tag is the exception: it must select nothing, because an
// uncovered edit fails the build.
func TestCoverageTags_EachSelectsEdits(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --format json"), &stdout, io.Discard))
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))

	for _, tag := range internal.GetCoverageTags() {
		count := 0
		for _, row := range rows {
			if row["status"] == tag || row["polarity"] == tag || row["action"] == tag {
				count++
			}
		}
		if tag == "uncovered" {
			require.Zero(t, count, "tag %q must select nothing while the coverage audit passes", tag)
		} else {
			require.Positive(t, count, "tag %q selects no edits", tag)
		}
	}
}

// Every tag must name exactly one dimension of the vocabulary; see the
// changelog listing's uniqueness test for why.
func TestCoverageTags_Unique(t *testing.T) {
	seen := map[string]bool{}
	for _, tag := range internal.GetCoverageTags() {
		require.False(t, seen[tag], "tag %q appears in two dimensions", tag)
		seen[tag] = true
	}
}

// the patterns listing has no edits to filter, so the flags are rejected
// together rather than one being silently ignored
func TestChecksCoverage_PatternsAndTagsRejected(t *testing.T) {
	var stdout, stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog coverage --patterns --tags covered"), &stdout, &stderr))
	require.Contains(t, stderr.String(), "--tags cannot be used with --patterns")
}
