//go:build unix

package internal

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

func TestReadSpecSource_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.yaml")
	body := []byte("openapi: 3.0.0\ninfo: {title: t, version: '1'}\npaths: {}\n")
	require.NoError(t, os.WriteFile(path, body, 0644))

	bytesOut, name, err := readSpecSource(load.NewSource(path))
	require.NoError(t, err)
	require.Equal(t, body, bytesOut)
	require.Equal(t, "openapi.yaml", name, "the upload's display filename must be the basename, not the full local path")
}

func TestReadSpecSource_FromGitRef(t *testing.T) {
	dir := t.TempDir()
	gitRun := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	gitRun("git", "init")
	gitRun("git", "config", "user.email", "t@t.com")
	gitRun("git", "config", "user.name", "t")
	body := []byte("openapi: 3.0.0\ninfo: {title: t, version: '1'}\npaths: {}\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), body, 0644))
	gitRun("git", "add", "openapi.yaml")
	gitRun("git", "commit", "-m", "init")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	bytesOut, name, err := readSpecSource(load.NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, body, bytesOut)
	require.Equal(t, "openapi.yaml", name, "display filename must be derived from the path portion of the git ref, not the ref itself")
}

func TestReadSpecSource_StdinRejected(t *testing.T) {
	_, _, err := readSpecSource(load.NewSource("-"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "stdin")
}

// decryptBlob is the test-side mirror of the browser decryptor: it reverses
// the version(1) || nonce(12) || ciphertext+tag layout using the key the CLI
// returned. If this round-trips, a correct WebCrypto implementation will too.
func decryptBlob(t *testing.T, blob, key []byte) []byte {
	t.Helper()
	require.Equal(t, byte(encryptedReviewBlobVersion), blob[0], "first byte must be the format version")
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := blob[1 : 1+gcm.NonceSize()]
	ciphertext := blob[1+gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	require.NoError(t, err, "ciphertext must decrypt with the returned key")
	return plaintext
}

func TestEncryptReviewPayload_RoundTrip(t *testing.T) {
	plaintext := []byte(`{"hello":"world","spec":"openapi: 3.0.0"}`)
	blob, key, err := encryptReviewPayload(plaintext)
	require.NoError(t, err)
	require.Len(t, key, 32, "AES-256 needs a 256-bit key")

	// The blob must not contain the plaintext anywhere — it is ciphertext.
	require.NotContains(t, string(blob), "openapi: 3.0.0")

	got := decryptBlob(t, blob, key)
	require.Equal(t, plaintext, got)

	// Two encryptions of the same plaintext must differ (fresh key + nonce),
	// so a server seeing two identical specs can't tell they're identical.
	blob2, key2, err := encryptReviewPayload(plaintext)
	require.NoError(t, err)
	require.NotEqual(t, key, key2, "each upload must use a fresh key")
	require.NotEqual(t, blob, blob2, "ciphertext must not be deterministic")
}

func TestPostEncryptedReview_PostsBlobAnonymously(t *testing.T) {
	var (
		gotContentType string
		gotAuth        string
		gotBody        []byte
	)
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/encrypted-review", r.URL.Path)
		gotContentType = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"review_id":"abc-123","expires_at":1700000000}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)

	blob := []byte{encryptedReviewBlobVersion, 9, 9, 9}
	id, expires, err := postEncryptedReview(blob)
	require.NoError(t, err)
	require.Equal(t, "abc-123", id)
	require.Equal(t, int64(1700000000), expires.Unix())
	require.Equal(t, "application/octet-stream", gotContentType)
	require.Empty(t, gotAuth, "the encrypted upload is anonymous; no Authorization header is sent")
	require.Equal(t, blob, gotBody, "the raw ciphertext blob must be the request body")
}

func TestPostEncryptedReview_ServerError(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(`{"error":"too large"}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)

	_, _, err := postEncryptedReview([]byte{1, 2, 3})
	require.Error(t, err)
	require.Contains(t, err.Error(), "413")
}

// keyFragmentRe extracts the base64url key from a #k= fragment.
var keyFragmentRe = regexp.MustCompile(`#k=([A-Za-z0-9_-]+)`)

func TestUploadAndOpen_EncryptsSpecsAndEmitsFragmentURL(t *testing.T) {
	// End-to-end: stand up a stub /api/encrypted-review, run uploadAndOpen
	// against two real spec files, capture the uploaded blob, pull the key
	// out of the emitted #fragment, decrypt, and assert the payload carries
	// both specs verbatim and the right mode. This is the zero-knowledge
	// contract: the server only ever sees the opaque blob; only the
	// fragment key (which never leaves the URL) unlocks it.
	var uploadedBlob []byte
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/encrypted-review", r.URL.Path)
		uploadedBlob, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"review_id":"r-1","expires_at":0}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)

	baseDir := t.TempDir()
	basePath := filepath.Join(baseDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("openapi: 3.0.0\nbase\n"), 0644))
	revDir := t.TempDir()
	revPath := filepath.Join(revDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(revPath, []byte("openapi: 3.0.0\nrev\n"), 0644))

	flags := NewFlags()
	flags.setBase(load.NewSource(basePath))
	flags.setRevision(load.NewSource(revPath))

	var out bytes.Buffer
	// nil specInfoPair is safe (version getters return "n/a"); empty changes
	// is a valid changelog.
	require.NoError(t, uploadAndOpen(flags, &out, false, checker.Changes{}, nil, true))

	// The emitted URL must point at the encrypted review surface and carry
	// the key only in the fragment.
	require.Contains(t, out.String(), "/review/e/r-1#k=")
	m := keyFragmentRe.FindStringSubmatch(out.String())
	require.Len(t, m, 2, "the URL must contain a #k= fragment")
	key, err := base64.RawURLEncoding.DecodeString(m[1])
	require.NoError(t, err)
	require.Len(t, key, 32)

	// The server-visible blob must not leak the spec content in cleartext.
	require.NotContains(t, string(uploadedBlob), "base")
	require.NotContains(t, string(uploadedBlob), "rev")

	var payload reviewPayload
	require.NoError(t, json.Unmarshal(decryptBlob(t, uploadedBlob, key), &payload))
	require.Equal(t, "openapi: 3.0.0\nbase\n", payload.BaseSpec, "base spec must round-trip verbatim")
	require.Equal(t, "openapi: 3.0.0\nrev\n", payload.RevisionSpec, "revision spec must round-trip verbatim")
	require.Equal(t, "openapi.yaml", payload.BaseFilename)
	require.Equal(t, "openapi.yaml", payload.RevisionFilename)
	require.Equal(t, "changelog", payload.Mode, "isBreaking=false maps to mode=changelog")
	require.NotEmpty(t, payload.Changes, "the computed changelog must be embedded so the page needs no recompute")
}

func TestUploadAndOpen_BreakingSetsModeBreaking(t *testing.T) {
	var uploadedBlob []byte
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadedBlob, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"review_id":"r2","expires_at":0}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)

	dir := t.TempDir()
	basePath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("openapi: 3.0.0\n"), 0644))
	revDir := t.TempDir()
	revPath := filepath.Join(revDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(revPath, []byte("openapi: 3.0.0\n"), 0644))

	flags := NewFlags()
	flags.setBase(load.NewSource(basePath))
	flags.setRevision(load.NewSource(revPath))

	var out bytes.Buffer
	require.NoError(t, uploadAndOpen(flags, &out, true, checker.Changes{}, nil, true))

	key, err := base64.RawURLEncoding.DecodeString(keyFragmentRe.FindStringSubmatch(out.String())[1])
	require.NoError(t, err)
	var payload reviewPayload
	require.NoError(t, json.Unmarshal(decryptBlob(t, uploadedBlob, key), &payload))
	require.Equal(t, "breaking", payload.Mode, "isBreaking=true must propagate as mode=breaking")
}
