package internal_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Guards are output and queryable: --tags read-only selects exactly the rules
// declaring the guard, each row carries its guards, and the id naming
// convention stops being load-bearing for the audit.
func Test_ChecksChangelogGuards(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --tags read-only"), &stdout, io.Discard))

	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	require.NotEmpty(t, checks)
	for _, check := range checks {
		require.Contains(t, check["guards"], "read-only", check["id"])
	}
}

// The text format renders the guards column.
func Test_ChecksChangelogGuardsTextColumn(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --tags sanctioned"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "GUARDS")
	require.Contains(t, stdout.String(), "sanctioned")
}

// --id displays exactly the named check, with its full record in json.
func Test_ChecksChangelogIdFilter(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --id api-removed-without-deprecation"), &stdout, io.Discard))

	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	require.Len(t, checks, 1)
	require.Equal(t, "api-removed-without-deprecation", checks[0]["id"])
	require.NotEmpty(t, checks[0]["locations"])
}

func Test_ChecksChangelogUnknownIdRejected(t *testing.T) {
	var stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog --id no-such-check"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), `unknown check id "no-such-check"`)
}

// --location keeps only checks claiming a location containing the string.
func Test_ChecksChangelogLocationFilter(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --location requestBody.content.*.schema.maximum"), &stdout, io.Discard))

	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	require.NotEmpty(t, checks)
	for _, check := range checks {
		require.Contains(t, fmt.Sprint(check["locations"]), "requestBody.content.*.schema.maximum", check["id"])
	}
}

// The generated and hand-written tags partition the catalog, and the json
// generated field agrees with the tag that selected the check.
func Test_ChecksChangelogGeneratedPartition(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json --tags generated"), &stdout, io.Discard))
	var generated []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &generated))
	require.NotEmpty(t, generated)
	for _, check := range generated {
		require.Equal(t, true, check["generated"], check["id"])
	}

	handWritten := countRows(t, "oasdiff checks changelog --format json --tags hand-written")
	require.Positive(t, handWritten)
	require.Equal(t, countRows(t, "oasdiff checks changelog --format json"), len(generated)+handWritten)
}
