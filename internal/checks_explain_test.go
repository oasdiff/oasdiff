package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func explainOut(t *testing.T, cmd string) string {
	t.Helper()
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs(cmd), &stdout, io.Discard))
	return stdout.String()
}

// A changelog check explains its severity as the law's derivation.
func Test_ChecksExplainChangelogRule(t *testing.T) {
	out := explainOut(t, "oasdiff checks changelog explain api-removed-without-deprecation")
	require.Contains(t, out, "api-removed-without-deprecation  error")
	require.Contains(t, out, "Severity: error, derived.")
	require.Contains(t, out, "narrows")
	require.Contains(t, out, "Locations: paths.*.*:remove")
	require.Contains(t, out, `--severity-levels`)
}

// A guarded check's reasoning names what the guard did.
func Test_ChecksExplainGuardReasoning(t *testing.T) {
	out := explainOut(t, "oasdiff checks changelog explain request-read-only-property-max-decreased")
	require.Contains(t, out, "Severity: info, derived.")
	require.Contains(t, out, "read-only, so it never appears in requests")
}

// explain is scoped to the changelog checks, like the listing it sits under,
// so a validate id is unknown to it.
func Test_ChecksExplainRejectsValidateId(t *testing.T) {
	var stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog explain additional-operations-duplicate-method"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), `unknown check id "additional-operations-duplicate-method"`)
}

// The listing's flags are local to `checks changelog`, so a subcommand that
// does not declare one rejects it instead of parsing and ignoring it: on
// explain, --id would read as a second check to explain.
func Test_ChecksChangelogSubcommandsRejectListingFlags(t *testing.T) {
	for cmd, flags := range map[string][]string{
		// explain declares its own --format and --lang
		"oasdiff checks changelog explain api-removed-without-deprecation": {"--id api-removed-without-deprecation", "--location paths", "--tags request", "--severity error"},
		// coverage declares its own --format, --id, --location and --tags
		"oasdiff checks changelog coverage": {"--severity error", "--lang es"},
	} {
		for _, flag := range flags {
			var stderr bytes.Buffer
			require.NotZero(t, internal.Run(cmdToArgs(cmd+" "+flag), io.Discard, &stderr), cmd+" "+flag)
			require.Contains(t, stderr.String(), "unknown flag", cmd+" "+flag)
		}
	}
}

// The json record carries the derivation so tooling can embed it.
func Test_ChecksExplainJson(t *testing.T) {
	out := explainOut(t, "oasdiff checks changelog explain request-body-max-set --format json")

	var e map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &e))
	require.Equal(t, "request-body-max-set", e["id"])
	require.Equal(t, "error", e["level"])
	require.Equal(t, true, e["derived"])
	require.NotEmpty(t, e["reasoning"])
	require.NotEmpty(t, e["locations"])
}

func Test_ChecksExplainUnknownIdRejected(t *testing.T) {
	var stderr bytes.Buffer
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog explain no-such-check"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), `unknown check id "no-such-check"`)
}

func Test_ChecksExplainRequiresExactlyOneId(t *testing.T) {
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog explain"), io.Discard, io.Discard))
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks changelog explain a b"), io.Discard, io.Discard))
}

// Every changelog check can be explained: the id resolves and the explanation
// carries the level the listing shows.
func Test_ChecksExplainCoversEveryId(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff checks changelog --format json"), &stdout, io.Discard))
	var checks []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
	for _, check := range checks {
		id := check["id"].(string)
		out := explainOut(t, "oasdiff checks changelog explain --format json "+id)
		var e map[string]any
		require.NoError(t, json.Unmarshal([]byte(out), &e), id)
		require.Equal(t, check["level"], e["level"], id)
	}
}
