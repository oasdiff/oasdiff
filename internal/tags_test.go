package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func Test_ChecksNoTags(t *testing.T) {
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog -l ru"), io.Discard, io.Discard))
}

// countRows runs the command and returns how many rows its json output holds.
func countRows(t *testing.T, cmd string) int {
	t.Helper()
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs(cmd), &stdout, io.Discard))
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	return len(rows)
}

// The tags are taken from the vocabulary itself, so a tag added there is
// tested here without anyone remembering to. Every tag must select at
// least one row: a tag selecting nothing is a dead vocabulary entry or a
// broken matcher.
func TestChangelogTags_EachSelectsRules(t *testing.T) {
	for _, tag := range internal.GetChangelogTagsForTest() {
		require.Positive(t, countRows(t, "oasdiff checks changelog --format json --tags "+tag), "tag %q selects no rules", tag)
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

	for _, tag := range internal.GetCoverageTagsForTest() {
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

// Every tag must name exactly one dimension of its vocabulary: filtering
// groups requested values by dimension, so a value appearing in two
// dimensions would be ambiguous. A future dimension whose natural value
// collides must pick another name.
func TestTagVocabulariesUnique(t *testing.T) {
	for name, tags := range map[string][]string{
		"checks changelog":          internal.GetChangelogTagsForTest(),
		"checks changelog coverage": internal.GetCoverageTagsForTest(),
	} {
		seen := map[string]bool{}
		for _, tag := range tags {
			require.False(t, seen[tag], "%s: tag %q appears in two dimensions", name, tag)
			seen[tag] = true
		}
	}
}

// Values of the same dimension are ORed, dimensions are ANDed.
func TestTags_OrWithinDimension(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --tags request,response,decrease"), &stdout, io.Discard))

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &rows))
	require.NotEmpty(t, rows)
	sawRequest, sawResponse := false, false
	for _, row := range rows {
		require.Contains(t, []any{"request", "response"}, row["direction"])
		require.Contains(t, row["actions"], "decrease")
		sawRequest = sawRequest || row["direction"] == "request"
		sawResponse = sawResponse || row["direction"] == "response"
	}
	require.True(t, sawRequest, "OR within the direction dimension must include request rows")
	require.True(t, sawResponse, "OR within the direction dimension must include response rows")
}
