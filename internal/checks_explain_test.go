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
	out := explainOut(t, "oasdiff checks explain api-removed-without-deprecation")
	require.Contains(t, out, "api-removed-without-deprecation  error")
	require.Contains(t, out, "Severity: error, derived.")
	require.Contains(t, out, "narrows")
	require.Contains(t, out, "Locations: paths.*.*:remove")
	require.Contains(t, out, `--severity-levels`)
}

// A guarded check's reasoning names what the guard did.
func Test_ChecksExplainGuardReasoning(t *testing.T) {
	out := explainOut(t, "oasdiff checks explain request-read-only-property-max-decreased")
	require.Contains(t, out, "Severity: info, derived.")
	require.Contains(t, out, "read-only, so it never appears in requests")
}

// A validate rule has no taxonomy: its severity is set, not derived.
func Test_ChecksExplainValidateRule(t *testing.T) {
	out := explainOut(t, "oasdiff checks explain additional-operations-duplicate-method")
	require.Contains(t, out, "Severity: error, set by the rule.")
	require.NotContains(t, out, "Scope:")
}

// The json record carries the derivation so tooling can embed it.
func Test_ChecksExplainJson(t *testing.T) {
	out := explainOut(t, "oasdiff checks explain request-body-max-set --format json")

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
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks explain no-such-check"), io.Discard, &stderr))
	require.Contains(t, stderr.String(), `unknown check id "no-such-check"`)
}

func Test_ChecksExplainRequiresExactlyOneId(t *testing.T) {
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks explain"), io.Discard, io.Discard))
	require.NotZero(t, internal.Run(cmdToArgs("oasdiff checks explain a b"), io.Discard, io.Discard))
}

// Every check in both listings can be explained: the id resolves and the
// explanation carries a level, so no id a user can encounter is unexplained.
// The same sweep asserts the two listings share no id: explain resolves both
// rule sets in one namespace, so a changelog id would silently shadow a
// validate id with the same name.
func Test_ChecksExplainCoversEveryId(t *testing.T) {
	seen := map[string]string{}
	for _, listing := range []string{"oasdiff checks changelog --format json", "oasdiff checks validate --format json"} {
		var stdout bytes.Buffer
		require.Zero(t, internal.Run(cmdToArgs(listing), &stdout, io.Discard))
		var checks []map[string]any
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &checks))
		for _, check := range checks {
			id := check["id"].(string)
			require.NotContains(t, seen, id, "id %q appears in both %q and %q: rename it, or `checks explain` cannot keep a single id namespace", id, seen[id], listing)
			seen[id] = listing
			out := explainOut(t, "oasdiff checks explain --format json "+id)
			var e map[string]any
			require.NoError(t, json.Unmarshal([]byte(out), &e), id)
			require.Equal(t, check["level"], e["level"], id)
		}
	}
}
