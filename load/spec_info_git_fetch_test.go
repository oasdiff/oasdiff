//go:build unix

package load_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

const specV1 = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`

const specV2 = `openapi: "3.0.0"
info:
  title: Test
  version: "2.0"
paths: {}
`

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return strings.TrimSpace(string(out))
}

// TestLoadInfo_GitRevisionFetch verifies the --fetch flag: a git-revision source
// whose commit is missing from the local clone is fetched from "origin" when
// Source.Fetch is true, and without it the load fails with the actionable hint
// that points at --fetch.
func TestLoadInfo_GitRevisionFetch(t *testing.T) {
	origin := t.TempDir()
	gitRun(t, origin, "git", "init")
	gitRun(t, origin, "git", "config", "user.email", "test@test.com")
	gitRun(t, origin, "git", "config", "user.name", "Test")
	// Let the local transport serve a specific commit by SHA on fetch.
	gitRun(t, origin, "git", "config", "uploadpack.allowAnySHA1InWant", "true")
	gitRun(t, origin, "git", "config", "uploadpack.allowReachableSHA1InWant", "true")

	require.NoError(t, os.WriteFile(filepath.Join(origin, "openapi.yaml"), []byte(specV1), 0644))
	gitRun(t, origin, "git", "add", "openapi.yaml")
	gitRun(t, origin, "git", "commit", "-m", "v1")

	// Clone while origin is at v1, so the clone will lack the next commit.
	clone := t.TempDir()
	gitRun(t, filepath.Dir(clone), "git", "clone", origin, clone)

	// Advance origin to v2 — the commit the clone is missing.
	require.NoError(t, os.WriteFile(filepath.Join(origin, "openapi.yaml"), []byte(specV2), 0644))
	gitRun(t, origin, "git", "add", "openapi.yaml")
	gitRun(t, origin, "git", "commit", "-m", "v2")
	v2sha := gitRun(t, origin, "git", "rev-parse", "HEAD")

	// Run from inside the clone, where v2 is not present.
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(clone))
	defer os.Chdir(oldDir) //nolint:errcheck

	require.Error(t, exec.Command("git", "cat-file", "-e", v2sha).Run(),
		"precondition: the clone must not already have v2")

	ref := v2sha + ":openapi.yaml"

	// Without --fetch: the commit is missing, so the load fails and the hint
	// points at --fetch.
	_, err = load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(ref))
	require.Error(t, err)
	require.Contains(t, err.Error(), "--fetch")

	// With --fetch: oasdiff fetches v2 from origin and loads it.
	src := load.NewSource(ref)
	src.Fetch = true
	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), src)
	require.NoError(t, err)
	require.Equal(t, "2.0", specInfo.GetVersion())

	// The fetch populated the local object store.
	require.NoError(t, exec.Command("git", "cat-file", "-e", v2sha).Run())
}

// TestLoadInfo_GitRevisionFetchNoOriginFails verifies that --fetch surfaces a
// clear error when there is no "origin" remote to fetch from (rather than
// silently falling through).
func TestLoadInfo_GitRevisionFetchNoOrigin(t *testing.T) {
	dir := t.TempDir()
	gitRun(t, dir, "git", "init")
	gitRun(t, dir, "git", "config", "user.email", "test@test.com")
	gitRun(t, dir, "git", "config", "user.name", "Test")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	// A plausible-looking but absent commit SHA.
	ref := "0123456789012345678901234567890123456789:openapi.yaml"
	src := load.NewSource(ref)
	src.Fetch = true
	_, err = load.NewSpecInfo(openapi3.NewLoader(), src)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fetch")
}
