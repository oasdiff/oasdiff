package internal_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

const baseSpec = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
`

// breakingSpec removes the /pets endpoint, a breaking change (ERR).
const breakingSpec = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`

// breakingFilesRepo creates a git repo committing the given files, then chdirs
// into it. Callers overwrite files in the working tree to stage changes for the
// breaking check to compare against HEAD.
func breakingFilesRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	gitRun := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	gitRun("git", "init")
	gitRun("git", "config", "user.email", "test@test.com")
	gitRun("git", "config", "user.name", "Test")

	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
		gitRun("git", "add", name)
	}
	gitRun("git", "commit", "-m", "base")

	// t.Chdir restores the previous directory automatically and fails the test
	// if it is parallel, so it can't quietly corrupt sibling tests.
	t.Chdir(dir)

	return dir
}

// Test_BreakingFiles_Fails compares a working-tree file that dropped an endpoint
// against its committed version and asserts --fail-on ERR blocks with exit 1.
func Test_BreakingFiles_Fails(t *testing.T) {
	dir := breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --fail-on ERR openapi.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// Test_BreakingFiles_Passes asserts an unchanged file is not a breaking change,
// so the command exits 0 even with --fail-on ERR.
func Test_BreakingFiles_Passes(t *testing.T) {
	breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})

	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --fail-on ERR openapi.yaml"), io.Discard, io.Discard)

	require.Zero(t, exitCode)
}

// Test_BreakingFiles_MultipleFiles checks every file and aggregates the result:
// one breaking file among several still fails the whole run, and both files'
// output is printed.
func Test_BreakingFiles_MultipleFiles(t *testing.T) {
	dir := breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec, "other.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --fail-on ERR openapi.yaml other.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	// both files reported: the breaking one names the removed path, and each
	// result is labelled with its filename so they can be told apart
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
	require.Contains(t, stdout.String(), "openapi.yaml")
	require.Contains(t, stdout.String(), "other.yaml")
}

// Test_BreakingFiles_NewFileSkipped checks that a file absent from the base ref
// (a newly added spec) is skipped rather than aborting the run: the new file is
// listed first, yet the breaking change in a later, existing file is still
// caught and fails the command.
func Test_BreakingFiles_NewFileSkipped(t *testing.T) {
	dir := breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))
	// new-api.yaml exists in the working tree but was never committed, so it has
	// no version in HEAD
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new-api.yaml"), []byte(baseSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --fail-on ERR new-api.yaml openapi.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "new file, not in base ref, skipped")
	require.Contains(t, stdout.String(), "new-api.yaml")
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// Test_BreakingFiles_NoFailOn prints breaking changes but keeps exit 0 when
// --fail-on isn't set, matching the non-base breaking command.
func Test_BreakingFiles_NoFailOn(t *testing.T) {
	dir := breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD openapi.yaml"), &stdout, io.Discard)

	require.Zero(t, exitCode)
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// A path in neither the base ref nor the working tree is a mistake, not a new
// file. Reporting it as newly added would exit 0 and tell the caller their
// specs are clean when nothing was compared.
func Test_BreakingFiles_MissingFileIsNotNew(t *testing.T) {
	breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --fail-on ERR does-not-exist.yaml"), io.Discard, &stderr)

	require.NotZero(t, exitCode, "a path that exists nowhere must not pass")
	require.NotContains(t, stderr.String(), "skipped")
}

// Each positional is joined to the base ref as "<ref>:<path>", so only a file
// path makes sense. The others are rejected by what they are, not one at a
// time: stdin and a URL have no version in the ref, and a git revision would
// be joined into "HEAD:HEAD:openapi.yaml".
func Test_BreakingFiles_RejectsNonFileArguments(t *testing.T) {
	for name, arg := range map[string]string{
		"stdin":        "-",
		"url":          "https://example.com/openapi.yaml",
		"git revision": "HEAD:openapi.yaml",
	} {
		t.Run(name, func(t *testing.T) {
			breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})

			var stderr bytes.Buffer
			exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD "+arg), io.Discard, &stderr)

			require.Equal(t, 100, exitCode)
			require.Contains(t, stderr.String(), "is not a file path")
		})
	}
}

// breaking-files compares each spec on its own, so it has no second glob to
// merge and does not register --composed at all.
func Test_BreakingFiles_HasNoComposedFlag(t *testing.T) {
	breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --composed openapi.yaml"), io.Discard, &stderr)

	require.NotZero(t, exitCode)
	require.Contains(t, stderr.String(), "unknown flag: --composed")
}

// The flag-only validations are shared with the base+revision path, so a check
// added to checkCommonFlags reaches --base too. Pinned here because the two
// paths validate their positionals differently and could drift again.
func Test_BreakingFiles_SharesCommonFlagChecks(t *testing.T) {
	breakingFilesRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	// json renders no template, so checkTemplate rejects the pair. Templates
	// themselves work here, as they do on breaking and changelog; this is the
	// shared check firing, not a breaking-files limitation.
	exitCode := internal.Run(cmdToArgs("oasdiff breaking-files --base HEAD --format json --template x.tmpl openapi.yaml"), io.Discard, &stderr)

	require.NotZero(t, exitCode)
	require.Contains(t, stderr.String(), "template flag is not supported")
	require.NotContains(t, stderr.String(), "=== openapi.yaml ===", "must fail before any per-file output")
}
