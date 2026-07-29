package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// `oasdiff checks` names no rule set, so it has no listing of its own: it
// prints the help pointing at the two that do.
func Test_ChecksRequiresASubcommand(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks"), &stdout, io.Discard))

	out := stdout.String()
	require.Contains(t, out, "oasdiff checks [command]")
	require.Contains(t, out, "changelog")
	require.Contains(t, out, "validate")
	// No listing: the rules themselves are only reachable through a subcommand.
	require.NotContains(t, out, "api-deprecated-sunset-missing")
}

// A typo'd subcommand is a user error, so it has to be reported like one: a
// script redirecting the output would otherwise write help text to its target
// and see success.
func Test_ChecksUnknownSubcommandFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks chnagelog"), &stdout, &stderr))
	require.Contains(t, stderr.String(), `unknown command "chnagelog" for "oasdiff checks"`)
}

// The parent carries no listing flags, so an old `oasdiff checks --format json`
// fails loudly rather than quietly producing nothing for a caller piping it.
func Test_ChecksWithListingFlagsFails(t *testing.T) {
	for _, cmd := range []string{
		"oasdiff checks --format json",
		"oasdiff checks --severity error",
		"oasdiff checks --tags request",
	} {
		t.Run(cmd, func(t *testing.T) {
			require.NotZero(t, internal.Run(cmdToArgs(cmd), io.Discard, io.Discard))
		})
	}
}

func Test_ChecksSubcommandsListTheirRules(t *testing.T) {
	for _, cmd := range []string{
		"oasdiff checks changelog --format json",
		"oasdiff checks validate --format json",
	} {
		t.Run(cmd, func(t *testing.T) {
			var stdout bytes.Buffer
			require.Zero(t, internal.Run(cmdToArgs(cmd), &stdout, io.Discard))

			var checks []map[string]any
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
			require.NotEmpty(t, checks)
			for _, c := range checks {
				require.NotEmpty(t, c["id"])
				require.NotEmpty(t, c["level"], "every rule reports a severity")
			}
		})
	}
}

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
