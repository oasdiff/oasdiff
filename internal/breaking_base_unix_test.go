//go:build unix

package internal_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

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
	// both files reported: the breaking one names the removed path, the
	// unchanged one reports no breaking changes
	require.Contains(t, stdout.String(), "api-path-removed-without-deprecation")
	require.Equal(t, 2, strings.Count(stdout.String(), "changes"))
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
