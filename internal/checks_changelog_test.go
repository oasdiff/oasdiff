package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// The listing flags moved to the subcommand, so they have to work there.
func Test_ChecksChangelogAcceptsTheListingFlags(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(
		cmdToArgs("oasdiff checks changelog -l ru --tags decrease,parameters --severity info,warn,error --format json"),
		&stdout, io.Discard))

	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	require.NotEmpty(t, checks)
}

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
	for _, tag := range internal.GetChangelogTags() {
		require.Positive(t, countRows(t, "oasdiff checks changelog --format json --tags "+tag), "tag %q selects no rules", tag)
	}
}

// Every tag must name exactly one dimension of the vocabulary: filtering
// groups requested values by dimension, so a value appearing in two
// dimensions would be ambiguous. A future dimension whose natural value
// collides must pick another name.
func TestChangelogTags_Unique(t *testing.T) {
	seen := map[string]bool{}
	for _, tag := range internal.GetChangelogTags() {
		require.False(t, seen[tag], "tag %q appears in two dimensions", tag)
		seen[tag] = true
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
