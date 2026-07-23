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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/oasdiff/oasdiff/review"
	"github.com/stretchr/testify/require"
)

func TestReadSpecSource_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.yaml")
	body := []byte("openapi: 3.0.0\ninfo: {title: t, version: '1'}\npaths: {}\n")
	require.NoError(t, os.WriteFile(path, body, 0644))

	bytesOut, name, err := readSpecSource(load.NewSource(path), nil)
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

	bytesOut, name, err := readSpecSource(load.NewSource("HEAD:openapi.yaml"), nil)
	require.NoError(t, err)
	require.Equal(t, body, bytesOut)
	require.Equal(t, "openapi.yaml", name, "display filename must be derived from the path portion of the git ref, not the ref itself")
}

func TestReadSpecSource_StdinRejected(t *testing.T) {
	_, _, err := readSpecSource(load.NewSource("-"), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "stdin")
}

// decryptBlob is the test-side mirror of the browser decryptor: it reverses
// the version(1) || nonce(12) || ciphertext+tag layout using the key the CLI
// returned. If this round-trips, a correct WebCrypto implementation will too.
func decryptBlob(t *testing.T, blob, key []byte) []byte {
	t.Helper()
	require.Equal(t, byte(review.BlobVersion), blob[0], "first byte must be the format version")
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

	blob := []byte{review.BlobVersion, 9, 9, 9}
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

func TestInternalReview(t *testing.T) {
	require.False(t, internalReview(), "unset by default")
	t.Setenv("OASDIFF_INTERNAL", "0")
	require.False(t, internalReview(), "only \"1\" enables the marker")
	t.Setenv("OASDIFF_INTERNAL", "1")
	require.True(t, internalReview())
}

func TestPostEncryptedReview_InternalMarkerOnUpload(t *testing.T) {
	var gotQuery string
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/encrypted-review", r.URL.Path)
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"review_id":"abc-123","expires_at":1700000000}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)
	t.Setenv("OASDIFF_INTERNAL", "1")

	_, _, err := postEncryptedReview([]byte{review.BlobVersion, 9, 9, 9})
	require.NoError(t, err)
	require.Equal(t, "internal=1", gotQuery, "OASDIFF_INTERNAL tags the upload so it can be excluded from review stats")
}

// keyFragmentRe extracts the base64url key from a #k= fragment.
var keyFragmentRe = regexp.MustCompile(`#k=([A-Za-z0-9_-]+)`)

func TestUploadAndOpen_EncryptsSpecsAndEmitsFragmentURL(t *testing.T) {
	stubBrowser(t)
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
	require.NoError(t, uploadAndOpen(flags, &out, false, checker.Changes{}, nil, nil, true))

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

	var payload review.Payload
	require.NoError(t, json.Unmarshal(decryptBlob(t, uploadedBlob, key), &payload))
	require.Equal(t, "openapi: 3.0.0\nbase\n", payload.BaseSpec, "base spec must round-trip verbatim")
	require.Equal(t, "openapi: 3.0.0\nrev\n", payload.RevisionSpec, "revision spec must round-trip verbatim")
	require.Equal(t, "openapi.yaml", payload.BaseFilename)
	require.Equal(t, "openapi.yaml", payload.RevisionFilename)
	require.Equal(t, "changelog", payload.Mode, "isBreaking=false maps to mode=changelog")
	require.NotEmpty(t, payload.Changes, "the computed changelog must be embedded so the page needs no recompute")
}

func TestUploadAndOpen_BreakingSetsModeBreaking(t *testing.T) {
	stubBrowser(t)
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
	require.NoError(t, uploadAndOpen(flags, &out, true, checker.Changes{}, nil, nil, true))

	key, err := base64.RawURLEncoding.DecodeString(keyFragmentRe.FindStringSubmatch(out.String())[1])
	require.NoError(t, err)
	var payload review.Payload
	require.NoError(t, json.Unmarshal(decryptBlob(t, uploadedBlob, key), &payload))
	require.Equal(t, "breaking", payload.Mode, "isBreaking=true must propagate as mode=breaking")
}

func TestParseReviewMeta(t *testing.T) {
	ok := func(t *testing.T, want map[string]string, entries ...string) {
		t.Helper()
		got, err := parseReviewMeta(entries)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
	bad := func(t *testing.T, wantMsg string, entries ...string) {
		t.Helper()
		_, err := parseReviewMeta(entries)
		require.Error(t, err)
		require.Contains(t, err.Error(), wantMsg)
	}

	t.Run("simple key=value", func(t *testing.T) {
		ok(t, map[string]string{"a": "b"}, "a=b")
	})
	t.Run("value containing = splits on the first only", func(t *testing.T) {
		ok(t, map[string]string{"k": "a=b=c"}, "k=a=b=c")
	})
	t.Run("multiple entries", func(t *testing.T) {
		ok(t, map[string]string{"a": "1", "b": "2"}, "a=1", "b=2")
	})
	t.Run("empty value is allowed", func(t *testing.T) {
		ok(t, map[string]string{"a": ""}, "a=")
	})
	t.Run("nil input yields empty map", func(t *testing.T) {
		ok(t, map[string]string{})
	})
	t.Run("entry without = is an error, not silently dropped", func(t *testing.T) {
		bad(t, "expected key=value", "a=b", "noequals")
	})
	t.Run("leading = (empty key) is an error", func(t *testing.T) {
		bad(t, "expected key=value", "=value")
	})
	t.Run("duplicate key is an error, not last-wins", func(t *testing.T) {
		bad(t, "duplicate", "a=first", "a=second")
	})
}

func TestUploadAuthenticatedReview_InvalidMetaIsError(t *testing.T) {
	// A malformed --review-meta entry must surface (the caller demotes it to a
	// warning), not be silently dropped. Fails during parsing, before any HTTP
	// call, so no server is needed.
	var out bytes.Buffer
	err := uploadAuthenticatedReview("tok", []string{"noequals"}, []byte{review.BlobVersion, 1}, make([]byte, 32), checker.Changes{}, &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected key=value")
}

// stubBrowser replaces openBrowser with a no-op for the test, so the upload
// tests don't actually launch a browser on the dev machine.
func stubBrowser(t *testing.T) {
	t.Helper()
	orig := openBrowser
	openBrowser = func(string) error { return nil }
	t.Cleanup(func() { openBrowser = orig })
}

// writeSpecPair writes two minimal specs to fresh temp dirs and returns a Flags
// pointed at them. The two dirs keep the basenames identical (openapi.yaml)
// without colliding, mirroring the existing --open tests.
func writeSpecPair(t *testing.T) *Flags {
	t.Helper()
	baseDir := t.TempDir()
	basePath := filepath.Join(baseDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("openapi: 3.0.0\nbase\n"), 0644))
	revDir := t.TempDir()
	revPath := filepath.Join(revDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(revPath, []byte("openapi: 3.0.0\nrev\n"), 0644))

	flags := NewFlags()
	flags.setBase(load.NewSource(basePath))
	flags.setRevision(load.NewSource(revPath))
	return flags
}

func TestUploadAndOpen_AuthenticatedPath(t *testing.T) {
	stubBrowser(t)
	var (
		gotPath        string
		gotMethod      string
		gotContentType string
		gotBody        []byte
	)
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		// A deliberately distinct host + the Pro /review/ep route: it proves the
		// CLI echoes the server-returned review_url verbatim, not OASDIFF_URL.
		_, _ = w.Write([]byte(`{"review_token":"rt-1","review_url":"https://review.example.test/review/ep/auth-1","gate":{"state":"approved","breaking_total":3,"breaking_approved":3}}`))
	}))
	defer stub.Close()
	// The authenticated upload targets the API base, not the site base.
	t.Setenv("OASDIFF_API_URL", stub.URL)
	// Point the site env elsewhere to prove the auth path does not use it.
	t.Setenv("OASDIFF_URL", "https://should-not-be-used.example.com")

	flags := writeSpecPair(t)
	flags.v.Set("review-token", "tok-123")
	flags.v.Set("review-meta", []string{"owner=acme", "pr=42"})

	var out bytes.Buffer
	require.NoError(t, uploadAndOpen(flags, &out, true, checker.Changes{}, nil, nil, true))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "/tenants/tok-123/encrypted-pr-review", gotPath, "the token is path-authenticated")
	require.Equal(t, "application/json", gotContentType)

	var body struct {
		Ciphertext []byte            `json:"ciphertext"`
		Metadata   map[string]string `json:"metadata"`
		Changes    []review.Change   `json:"changes"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &body))
	require.NotEmpty(t, body.Ciphertext, "the encrypted bundle must ride in ciphertext")
	require.Equal(t, byte(review.BlobVersion), body.Ciphertext[0], "ciphertext is the same blob layout as the free path")
	require.Equal(t, map[string]string{"owner": "acme", "pr": "42"}, body.Metadata, "the opaque metadata bag is forwarded verbatim")

	// The output carries the returned URL with the key in its #fragment and the
	// gate state printed verbatim.
	require.Contains(t, out.String(), "https://review.example.test/review/ep/auth-1#k=", "the CLI appends #k= to the server-returned review_url verbatim")
	require.Contains(t, out.String(), "oasdiff: review status: approved")
	require.NotContains(t, out.String(), "should-not-be-used", "the authenticated path must not use the site base URL")

	m := keyFragmentRe.FindStringSubmatch(out.String())
	require.Len(t, m, 2, "the URL must contain a #k= fragment")
	key, err := base64.RawURLEncoding.DecodeString(m[1])
	require.NoError(t, err)
	require.Len(t, key, 32)

	// The fragment key must decrypt the uploaded ciphertext: same zero-knowledge
	// contract as the free path.
	var payload review.Payload
	require.NoError(t, json.Unmarshal(decryptBlob(t, body.Ciphertext, key), &payload))
	require.Equal(t, "openapi: 3.0.0\nbase\n", payload.BaseSpec)
	require.Equal(t, "openapi: 3.0.0\nrev\n", payload.RevisionSpec)
}

func TestUploadAndOpen_FreePathWhenNoToken(t *testing.T) {
	stubBrowser(t)
	// With no token, --open must hit the free endpoint on the site base, not the
	// authenticated one. Fail loudly if the authenticated endpoint is touched.
	var freeHit bool
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/encrypted-review", r.URL.Path, "no token means the free anonymous endpoint")
		freeHit = true
		_, _ = w.Write([]byte(`{"review_id":"free-1","expires_at":0}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)
	// If the auth path were wrongly taken, this base would be used; point it at a
	// dead address so the test would error rather than silently pass.
	t.Setenv("OASDIFF_API_URL", "http://127.0.0.1:0")

	flags := writeSpecPair(t)

	var out bytes.Buffer
	require.NoError(t, uploadAndOpen(flags, &out, false, checker.Changes{}, nil, nil, true))
	require.True(t, freeHit, "the free endpoint must be used when --review-token is empty")
	require.Contains(t, out.String(), "/review/e/free-1#k=")
}

func TestUploadAuthenticatedReview_ServerErrorIsNonFatal(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"tenant not found"}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_API_URL", stub.URL)

	blob := []byte{review.BlobVersion, 1, 2, 3}
	key := make([]byte, 32)
	var out bytes.Buffer
	err := uploadAuthenticatedReview("tok", nil, blob, key, checker.Changes{}, &out)
	require.Error(t, err, "the function returns the error; the caller (getChangelog) demotes it to a warning")
	require.Contains(t, err.Error(), "403")
	require.Contains(t, err.Error(), "tenant not found", "the server's error body must surface in the warning")
}

func TestUploadAuthenticatedReview_UnreachableIsNonFatal(t *testing.T) {
	// Port 0 is not connectable; the upload must return an error (which the
	// caller demotes to a warning) rather than panicking.
	t.Setenv("OASDIFF_API_URL", "http://127.0.0.1:0")
	blob := []byte{review.BlobVersion, 1, 2, 3}
	var out bytes.Buffer
	err := uploadAuthenticatedReview("tok", nil, blob, make([]byte, 32), checker.Changes{}, &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "upload to")
}

func TestUploadAuthenticatedReview_InvalidJSONResponse(t *testing.T) {
	// A 2xx whose body isn't the expected JSON must surface a parse error, not a
	// bogus review URL.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_API_URL", stub.URL)

	var out bytes.Buffer
	err := uploadAuthenticatedReview("tok", nil, []byte{review.BlobVersion, 1}, make([]byte, 32), checker.Changes{}, &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse response")
	require.NotContains(t, out.String(), "Opening", "no URL is emitted when the response can't be parsed")
}

func TestUploadAuthenticatedReview_MissingReviewURL(t *testing.T) {
	// A 2xx that omits review_url can't produce a usable link, so it must error
	// rather than print a keyless or malformed URL.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"review_token":"rt","gate":{"state":"pending"}}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_API_URL", stub.URL)

	var out bytes.Buffer
	err := uploadAuthenticatedReview("tok", nil, []byte{review.BlobVersion, 1}, make([]byte, 32), checker.Changes{}, &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "review_url")
}

func TestUploadAndOpen_ReadBaseSpecError(t *testing.T) {
	// A missing base spec fails before any upload; the error must name the base.
	flags := NewFlags()
	flags.setBase(load.NewSource(filepath.Join(t.TempDir(), "does-not-exist.yaml")))
	flags.setRevision(load.NewSource(filepath.Join(t.TempDir(), "also-missing.yaml")))

	var out bytes.Buffer
	err := uploadAndOpen(flags, &out, false, checker.Changes{}, nil, nil, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "read base spec")
}

func TestUploadAndOpen_ReadRevisionSpecError(t *testing.T) {
	// Base reads fine but the revision is missing: the error must name the revision.
	dir := t.TempDir()
	basePath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("openapi: 3.0.0\n"), 0644))

	flags := NewFlags()
	flags.setBase(load.NewSource(basePath))
	flags.setRevision(load.NewSource(filepath.Join(t.TempDir(), "missing.yaml")))

	var out bytes.Buffer
	err := uploadAndOpen(flags, &out, false, checker.Changes{}, nil, nil, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "read revision spec")
}

// URL sources ride the capture: the bundle's specs are the exact bytes the
// loader fetched (no second fetch), and blocks slice from them.
func TestUploadAndOpen_URLSources(t *testing.T) {
	stubBrowser(t)
	const baseSpec = "openapi: 3.0.0\ninfo: {title: t, version: \"1\"}\npaths:\n  /users:\n    get:\n      responses:\n        \"200\": {description: ok}\n        \"404\": {description: gone}\n"
	const revSpec = "openapi: 3.0.0\ninfo: {title: t, version: \"1\"}\npaths:\n  /users:\n    get:\n      responses:\n        \"200\": {description: ok}\n"
	specs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/base/openapi.yaml" {
			_, _ = w.Write([]byte(baseSpec))
			return
		}
		_, _ = w.Write([]byte(revSpec))
	}))
	defer specs.Close()

	var uploadedBlob []byte
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadedBlob, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"review_id":"r-url","expires_at":0}`))
	}))
	defer stub.Close()
	t.Setenv("OASDIFF_URL", stub.URL)

	baseURL := specs.URL + "/base/openapi.yaml"
	revURL := specs.URL + "/rev/openapi.yaml"
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	s1, err := load.NewSpecInfoWithCapture(loader, load.NewSource(baseURL))
	require.NoError(t, err)
	s2, err := load.NewSpecInfoWithCapture(loader, load.NewSource(revURL))
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)
	require.NotEmpty(t, errs)

	flags := NewFlags()
	flags.setBase(load.NewSource(baseURL))
	flags.setRevision(load.NewSource(revURL))

	var out bytes.Buffer
	require.NoError(t, uploadAndOpen(flags, &out, false, errs, []*load.SpecInfo{s1}, []*load.SpecInfo{s2}, false))

	m := keyFragmentRe.FindStringSubmatch(out.String())
	require.Len(t, m, 2)
	key, err := base64.RawURLEncoding.DecodeString(m[1])
	require.NoError(t, err)

	var payload review.Payload
	require.NoError(t, json.Unmarshal(decryptBlob(t, uploadedBlob, key), &payload))
	require.Equal(t, baseSpec, payload.BaseSpec, "the bundle carries the bytes the loader fetched")
	require.Equal(t, revSpec, payload.RevisionSpec)
	require.Equal(t, "openapi.yaml", payload.BaseFilename)
	require.NotEmpty(t, payload.Blocks)
	require.NotEmpty(t, payload.Blocks[0].BaseText, "blocks slice from the captured URL text")
}
