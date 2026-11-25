package internal_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

func cmdToArgs(cmd string) []string {
	return strings.Fields(cmd)
}

func TestMain(m *testing.M) {
	// No test may reach production by default: --open uploads go to these URLs
	// when a test forgets to stub them (it happened). Tests with a stub server
	// override per-test via t.Setenv.
	_ = os.Setenv("OASDIFF_URL", "http://127.0.0.1:0")
	_ = os.Setenv("OASDIFF_API_URL", "http://127.0.0.1:0")
	code := m.Run()
	os.Exit(code)
}

func Test_InvalidCmd(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff invalid"), io.Discard, io.Discard))
}

func Test_NoRevision(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff base.yaml"), io.Discard, io.Discard))
}

func Test_ExtraRevision(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff 1 2 3"), io.Discard, &stderr))
	require.Equal(t, "Error: invalid arguments after base and revision\n", stderr.String())
}

func Test_InvalidFlag(t *testing.T) {
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff data/openapi-test1.yaml data/openapi-test1.yaml --invalid"), io.Discard, io.Discard))
}

func Test_OpenWithComposedAllowed(t *testing.T) {
	// A composed diff (-c) can now be reviewed with --open: the cards carry the
	// composed comparison, so composed no longer conflicts with --open. Point the
	// free upload (OASDIFF_URL, not the authenticated OASDIFF_API_URL) at a
	// non-connectable address so no real network call happens; --open upload
	// errors are additive (non-fatal), so the exit code is the changelog result
	// (0 for identical specs), never the 100 arg-validation code.
	t.Setenv("OASDIFF_URL", "http://127.0.0.1:0")
	for _, cmd := range []string{
		"oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test1.yaml --composed --open",
		"oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test1.yaml -c --open",
	} {
		var stderr bytes.Buffer
		require.Equal(t, 0, internal.Run(cmdToArgs(cmd), io.Discard, &stderr), cmd)
	}
}

func Test_ReviewFlagsRequireOpen(t *testing.T) {
	// --review-token / --review-meta only make sense with --open; using either
	// without --open is rejected at argument validation before any diff runs.
	for _, cmd := range []string{
		"oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test1.yaml --review-token tok",
		"oasdiff breaking ../data/openapi-test1.yaml ../data/openapi-test1.yaml --review-meta owner=acme",
	} {
		var stderr bytes.Buffer
		require.Equal(t, 100, internal.Run(cmdToArgs(cmd), io.Discard, &stderr), cmd)
		require.Contains(t, stderr.String(), "--review-token and --review-meta require --open", cmd)
	}
}

func Test_ReviewFlagsAllowedWithOpen(t *testing.T) {
	// With --open the review flags pass argument validation. Point the upload at a
	// non-connectable API so no real network call happens; --open upload errors
	// are additive (warned, non-fatal), so the exit code is the changelog result
	// (0 for identical specs), never the 100 arg-validation code.
	t.Setenv("OASDIFF_API_URL", "http://127.0.0.1:0")
	var stderr bytes.Buffer
	code := internal.Run(cmdToArgs("oasdiff changelog ../data/openapi-test1.yaml ../data/openapi-test1.yaml --open --review-token tok"), io.Discard, &stderr)
	require.NotEqual(t, 100, code, "review flags with --open must pass argument validation")
	require.NotContains(t, stderr.String(), "require --open")
}

func Test_Help(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff --help"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "Usage:")
}

func Test_HelpShortcut(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff -h"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "Usage:")
}

func Test_Version(t *testing.T) {
	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff -v"), &stdout, io.Discard))
	require.Contains(t, stdout.String(), "oasdiff version")
}

func Test_InvalidEnumValue(t *testing.T) {
	var stderr bytes.Buffer
	require.Equal(t, 100, internal.Run(cmdToArgs("oasdiff diff ../data/openapi-test1.yaml ../data/openapi-test3.yaml --exclude-elements xxx"), io.Discard, &stderr))
	require.Equal(t, `Error: invalid argument "xxx" for "-e, --exclude-elements" flag: xxx is not one of the allowed values: description, endpoints, examples, extensions, summary, tags, or title
`, stderr.String())
}

// Two "-" operands name the same document (as in diff, where both stand for
// the same file): the single stdin read serves both sides, and the revision
// side carries the document's version, not an empty second read's ("n/a").
func Test_BothSidesStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(path, []byte("openapi: 3.0.0\ninfo: {title: t, version: \"1.2.3\"}\npaths: {}\n"), 0644))
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck
	orig := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = orig })

	var stdout bytes.Buffer
	require.Zero(t, internal.Run(cmdToArgs("oasdiff changelog - - --format html"), &stdout, io.Discard))
	require.Equal(t, 2, strings.Count(stdout.String(), "1.2.3"), "both sides carry the document's version")
	require.NotContains(t, stdout.String(), "n/a")
}
