package load

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests pin the contract the whole review-bundle depends on: a file the
// loader captured must be findable by the File on an element's origin, for every
// source type and on every platform. They exercise the same lookup
// review.Extract uses (by origin File, with a "./" fallback), so a regression in
// either the capture key or origin derivation fails here.

// textForOrigin mirrors review.Extract's lookup: by the origin File, then with
// the leading "./" trimmed (url.String() prepends it to colon-first git paths).
func textForOrigin(sources map[string]string, originFile string) (string, bool) {
	if s, ok := sources[originFile]; ok {
		return s, true
	}
	s, ok := sources[strings.TrimPrefix(originFile, "./")]
	return s, ok
}

func sourceKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// refElementOriginFile returns the origin File of the $ref'd User schema that
// captureRoot points at (defined in capture_test.go).
func refElementOriginFile(t *testing.T, si *SpecInfo) string {
	t.Helper()
	user := si.Spec.Paths.Value("/x").Get.Responses.Value("200").Value.Content["application/json"].Schema
	require.NotNil(t, user.Value.Origin, "the $ref'd element carries an origin")
	return user.Value.Origin.Key.File
}

func assertRefTextFindable(t *testing.T, si *SpecInfo) {
	t.Helper()
	originFile := refElementOriginFile(t, si)
	text, ok := textForOrigin(si.Sources, originFile)
	require.Truef(t, ok, "$ref'd file not findable by origin File %q; captured keys: %v", originFile, sourceKeys(si.Sources))
	require.Equal(t, captureSubDoc, text, "found text is the $ref'd file's content")
}

func TestCaptureFoundByOriginFile_LocalFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(captureRoot), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.yaml"), []byte(captureSubDoc), 0644))

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource(filepath.Join(dir, "openapi.yaml")))
	require.NoError(t, err)
	assertRefTextFindable(t, si)
}

func TestCaptureFoundByOriginFile_URL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(captureRoot)) })
	mux.HandleFunc("/other.yaml", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(captureSubDoc)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource(srv.URL+"/openapi.yaml"))
	require.NoError(t, err)
	assertRefTextFindable(t, si)
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
	assertRefTextFindable(t, si)
}
