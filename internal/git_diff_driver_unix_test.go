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

// nullBlob is the SHA-1 sentinel git passes as the blob hash for the added /
// deleted case in steady-state diffs. We pair it with "/dev/null" as the file
// path because that's what git uses end-to-end (verified empirically against
// git's external-diff protocol on the v2.46 release).
const nullBlob = "0000000000000000000000000000000000000000"

const minimalSpecV1 = `openapi: "3.0.0"
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

const minimalSpecV2 = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      responses:
        "200":
          description: ok
        "404":
          description: not found
`

// gitDiffDriverRepo sets up a fresh git repo containing two committed
// versions of openapi.yaml and returns the (dir, oldBlobHex, newBlobHex)
// triple plus a cleanup that restores the original working directory.
func gitDiffDriverRepo(t *testing.T) (dir, oldHex, newHex string) {
	t.Helper()
	dir = t.TempDir()

	gitRun := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}

	gitRun("git", "init")
	gitRun("git", "config", "user.email", "test@test.com")
	gitRun("git", "config", "user.name", "Test")

	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(minimalSpecV1), 0644))
	gitRun("git", "add", "openapi.yaml")
	gitRun("git", "commit", "-m", "v1")
	oldHex = gitRun("git", "rev-parse", "HEAD:openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(minimalSpecV2), 0644))
	gitRun("git", "add", "openapi.yaml")
	gitRun("git", "commit", "-m", "v2")
	newHex = gitRun("git", "rev-parse", "HEAD:openapi.yaml")

	require.NotEqual(t, oldHex, newHex)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	return dir, oldHex, newHex
}

// Test_GitDiffDriver_AddsNewResponse runs git-diff-driver in a real repo where
// the revision adds a "404" response and asserts the changelog output describes
// the change. This is the end-to-end happy path Jamie Tanna's blog post
// (jvt.me/posts/2026/04/11/oasdiff-driver) advocated for as a native feature.
func Test_GitDiffDriver_AddsNewResponse(t *testing.T) {
	_, oldHex, newHex := gitDiffDriverRepo(t)

	var stdout bytes.Buffer
	exitCode := internal.Run(
		[]string{"oasdiff", "git-diff-driver", "openapi.yaml", "/tmp/old", oldHex, "100644", "/tmp/new", newHex, "100644"},
		&stdout, io.Discard,
	)

	require.Zero(t, exitCode, "git-diff-driver must always exit 0 so it doesn't abort git's diff pipeline")
	require.NotEmpty(t, stdout.String(), "changelog output should be non-empty when responses differ")
	require.NotContains(t, stdout.String(), "/tmp/", "source labels must not leak temp file paths")
}

// Test_GitDiffDriver_Added asserts that when git passes the null-blob sentinel
// for the old hash (file added on this commit), the driver prints a one-line
// "Added <path>" notice rather than attempting a diff.
func Test_GitDiffDriver_Added(t *testing.T) {
	_, _, newHex := gitDiffDriverRepo(t)

	var stdout bytes.Buffer
	exitCode := internal.Run(
		[]string{"oasdiff", "git-diff-driver", "openapi.yaml", "/dev/null", nullBlob, "0", "/tmp/new", newHex, "100644"},
		&stdout, io.Discard,
	)

	require.Zero(t, exitCode)
	require.Equal(t, "Added openapi.yaml\n", stdout.String())
}

// Test_GitDiffDriver_Removed mirrors the Added case for deletions.
func Test_GitDiffDriver_Removed(t *testing.T) {
	_, oldHex, _ := gitDiffDriverRepo(t)

	var stdout bytes.Buffer
	exitCode := internal.Run(
		[]string{"oasdiff", "git-diff-driver", "openapi.yaml", "/tmp/old", oldHex, "100644", "/dev/null", nullBlob, "0"},
		&stdout, io.Discard,
	)

	require.Zero(t, exitCode)
	require.Equal(t, "Removed openapi.yaml\n", stdout.String())
}

// Test_GitDiffDriver_ModeOnlyChange asserts that when the two blob hashes are
// equal (mode-only change), the driver prints nothing — git's own machinery
// already surfaces the mode delta and oasdiff has nothing meaningful to add.
func Test_GitDiffDriver_ModeOnlyChange(t *testing.T) {
	_, oldHex, _ := gitDiffDriverRepo(t)

	var stdout bytes.Buffer
	exitCode := internal.Run(
		[]string{"oasdiff", "git-diff-driver", "openapi.yaml", "/tmp/old", oldHex, "100644", "/tmp/new", oldHex, "100755"},
		&stdout, io.Discard,
	)

	require.Zero(t, exitCode)
	require.Empty(t, stdout.String())
}

// Test_GitDiffDriver_NonGitRepo asserts that even when invoked outside any git
// repository the driver returns exit code 0 and writes the error inline. Aborting
// git's diff pipeline (non-zero exit) is the worst failure mode here because a
// single broken file would make `git log --ext-diff` unusable for the whole repo.
func Test_GitDiffDriver_NonGitRepo(t *testing.T) {
	tmp := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	defer os.Chdir(oldDir) //nolint:errcheck

	// Hand-rolled but plausible-looking hashes that don't resolve in this empty repo.
	const fakeOld = "1111111111111111111111111111111111111111"
	const fakeNew = "2222222222222222222222222222222222222222"

	var stdout bytes.Buffer
	exitCode := internal.Run(
		[]string{"oasdiff", "git-diff-driver", "openapi.yaml", "/tmp/old", fakeOld, "100644", "/tmp/new", fakeNew, "100644"},
		&stdout, io.Discard,
	)

	require.Zero(t, exitCode)
	require.Contains(t, stdout.String(), "oasdiff:", "error must be surfaced inline so the user sees it in git's diff output")
}
