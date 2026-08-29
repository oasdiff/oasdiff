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

// breakingBaseRepo creates a git repo committing the given files, then chdirs
// into it. Callers overwrite files in the working tree to stage changes for the
// breaking check to compare against HEAD.
func breakingBaseRepo(t *testing.T, files map[string]string) string {
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

// Test_BreakingBase_Fails compares a working-tree file that dropped an endpoint
// against its committed version and asserts --fail-on ERR blocks with exit 1.
func Test_BreakingBase_Fails(t *testing.T) {
	dir := breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --fail-on ERR openapi.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// Test_BreakingBase_Passes asserts an unchanged file is not a breaking change,
// so the command exits 0 even with --fail-on ERR.
func Test_BreakingBase_Passes(t *testing.T) {
	breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})

	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --fail-on ERR openapi.yaml"), io.Discard, io.Discard)

	require.Zero(t, exitCode)
}

// Test_BreakingBase_MultipleFiles checks every file and aggregates the result:
// one breaking file among several still fails the whole run, and both files'
// output is printed.
func Test_BreakingBase_MultipleFiles(t *testing.T) {
	dir := breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec, "other.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --fail-on ERR openapi.yaml other.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	// both files reported: the breaking one names the removed path, and each
	// result is labelled with its filename so they can be told apart
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
	require.Contains(t, stdout.String(), "openapi.yaml")
	require.Contains(t, stdout.String(), "other.yaml")
}

// Test_BreakingBase_NewFileSkipped checks that a file absent from the base ref
// (a newly added spec) is skipped rather than aborting the run: the new file is
// listed first, yet the breaking change in a later, existing file is still
// caught and fails the command.
func Test_BreakingBase_NewFileSkipped(t *testing.T) {
	dir := breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))
	// new-api.yaml exists in the working tree but was never committed, so it has
	// no version in HEAD
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new-api.yaml"), []byte(baseSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --fail-on ERR new-api.yaml openapi.yaml"), &stdout, io.Discard)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "new file, not in base ref, skipped")
	require.Contains(t, stdout.String(), "new-api.yaml")
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// Test_BreakingBase_NoFailOn prints breaking changes but keeps exit 0 when
// --fail-on isn't set, matching the non-base breaking command.
func Test_BreakingBase_NoFailOn(t *testing.T) {
	dir := breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(breakingSpec), 0644))

	var stdout bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD openapi.yaml"), &stdout, io.Discard)

	require.Zero(t, exitCode)
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
}

// A path in neither the base ref nor the working tree is a mistake, not a new
// file. Reporting it as newly added would exit 0 and tell the caller their
// specs are clean when nothing was compared.
func Test_BreakingBase_MissingFileIsNotNew(t *testing.T) {
	breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --fail-on ERR does-not-exist.yaml"), io.Discard, &stderr)

	require.NotZero(t, exitCode, "a path that exists nowhere must not pass")
	require.NotContains(t, stderr.String(), "skipped")
}

// The positionals are working-tree files that also have a version in the base
// ref, so stdin has nothing to compare against and is rejected up front rather
// than treated as a filename.
func Test_BreakingBase_RejectsStdin(t *testing.T) {
	breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD -"), io.Discard, &stderr)

	require.Equal(t, 100, exitCode)
	require.Contains(t, stderr.String(), "can't read from stdin with --base")
}

// Composed mode merges several specs into one comparison, which is the opposite
// of comparing each file against its own version in the base ref.
func Test_BreakingBase_RejectsComposed(t *testing.T) {
	breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --composed openapi.yaml"), io.Discard, &stderr)

	require.Equal(t, 100, exitCode)
	require.Contains(t, stderr.String(), "can't use --composed with --base")
}

// The flag-only validations are shared with the base+revision path, so a check
// added to checkCommonFlags reaches --base too. Pinned here because the two
// paths validate their positionals differently and could drift again.
func Test_BreakingBase_SharesCommonFlagChecks(t *testing.T) {
	breakingBaseRepo(t, map[string]string{"openapi.yaml": baseSpec})

	var stderr bytes.Buffer
	exitCode := internal.Run(cmdToArgs("oasdiff breaking --base HEAD --format json --template x.tmpl openapi.yaml"), io.Discard, &stderr)

	require.NotZero(t, exitCode)
	require.Contains(t, stderr.String(), "template flag is not supported")
	require.NotContains(t, stderr.String(), "=== openapi.yaml ===", "must fail before any per-file output")
}
