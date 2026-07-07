package load

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests pin the contract the whole review-bundle depends on: a file the
// loader captured must be findable by the File on an element's origin, for every
// source type and on every platform. They exercise the same lookup
// review.Extract uses (a direct match on origin File), so a regression in either
// the capture key or origin derivation fails here.

func sourceKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// assertCapturedByOrigin is the single source of truth for the capture<->origin
// contract: both files captureRoot pulls in -- the root itself and the one
// $ref'd file -- must be captured under the exact key kin records as each
// element's origin.Key.File, so review.Extract's slice lookup
// (Sources[origin.File]) hits. Each is looked up directly (no normalization),
// and nothing else may be captured.
func assertCapturedByOrigin(t *testing.T, si *SpecInfo) {
	t.Helper()

	// root document: Info is defined in the root spec
	require.NotNil(t, si.Spec.Info.Origin, "the root carries an origin")
	rootFile := si.Spec.Info.Origin.Key.File
	root, ok := si.Sources[rootFile]
	require.Truef(t, ok, "root not captured by its origin File %q; captured keys: %v", rootFile, sourceKeys(si.Sources))
	require.Equal(t, captureRoot, root, "root text captured verbatim")

	// $ref'd document: the resolved User schema lives in the other file
	user := si.Spec.Paths.Value("/x").Get.Responses.Value("200").Value.Content["application/json"].Schema
	require.NotNil(t, user.Value.Origin, "the $ref'd element carries an origin")
	refFile := user.Value.Origin.Key.File
	sub, ok := si.Sources[refFile]
	require.Truef(t, ok, "$ref'd file not captured by its origin File %q; captured keys: %v", refFile, sourceKeys(si.Sources))
	require.Equal(t, captureSubDoc, sub, "$ref'd file text captured verbatim")

	require.Len(t, si.Sources, 2, "exactly the root and the one $ref'd file are captured")
}

func TestCaptureFoundByOriginFile_LocalFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(captureRoot), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.yaml"), []byte(captureSubDoc), 0644))

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource(filepath.Join(dir, "openapi.yaml")))
	require.NoError(t, err)
	assertCapturedByOrigin(t, si)
}

func TestCaptureFoundByOriginFile_URL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(captureRoot)) })
	mux.HandleFunc("/other.yaml", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(captureSubDoc)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource(srv.URL+"/openapi.yaml"))
	require.NoError(t, err)
	assertCapturedByOrigin(t, si)
}

func TestCaptureFoundByOriginFile_GitRevision(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(captureRoot), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.yaml"), []byte(captureSubDoc), 0644))
	run("git", "add", ".")
	run("git", "commit", "-m", "spec")
	// Remove from disk so $ref resolution must go through "git show", exercising
	// the "HEAD:" prefixed origin path (not a plain filesystem read).
	require.NoError(t, os.Remove(filepath.Join(dir, "openapi.yaml")))
	require.NoError(t, os.Remove(filepath.Join(dir, "other.yaml")))

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() { _ = os.Chdir(oldDir) }()

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)
	assertCapturedByOrigin(t, si)
}
