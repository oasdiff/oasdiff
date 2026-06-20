package internal

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
)

// oasdiffSiteURL is the base URL of the web product. Defaults to the canonical
// oasdiff.com; set the OASDIFF_URL env var to point --open at a different
// deployment -- a local dev server, or a self-hosted oasdiff. The target must
// be a full web deployment: --open uploads to its /api/encrypted-review route
// and the review renders at /review/e, so a bare API service (api.oasdiff.com)
// won't work.
func oasdiffSiteURL() string {
	if u := os.Getenv("OASDIFF_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://www.oasdiff.com"
}

// oasdiffAPIBaseURL is the backend API base, used only by the authenticated
// review upload (the free path posts to oasdiffSiteURL). Override with
// OASDIFF_API_URL; defaults to api.oasdiff.com.
func oasdiffAPIBaseURL() string {
	if u := os.Getenv("OASDIFF_API_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://api.oasdiff.com"
}

// encryptedReviewBlobVersion is the first byte of the uploaded blob. It lets
// the browser decryptor reject a format it doesn't understand instead of
// trying to decrypt garbage. Bump it only on an incompatible layout change.
const encryptedReviewBlobVersion = 1

// reviewPayload is the cleartext bundle the CLI encrypts and uploads. The
// server (oasdiff.com by default, or whatever OASDIFF_URL points at) never
// sees it in cleartext: the CLI AES-256-GCM-encrypts the JSON below with a
// fresh random key, uploads only the ciphertext, and puts the key in the
// review URL's #fragment (which browsers never send to a server). The browser
// decrypts and renders.
//
// BaseSpec / RevisionSpec hold each spec's bytes verbatim as a string. A YAML
// spec stays YAML text and a JSON spec stays JSON text -- this JSON object is
// only the envelope that bundles the several fields into one blob; it does not
// reformat or reparse the spec content.
//
// Changes is the JSON changelog the CLI already computed, embedded raw. The
// server can't recompute it (it can't read the specs), so the CLI ships it.
// It is byte-identical to what the service's /public/changelog returns for the
// plaintext path, so the review page renders it the same way.
type reviewPayload struct {
	BaseSpec         string          `json:"base_spec" yaml:"base_spec"`
	RevisionSpec     string          `json:"revision_spec" yaml:"revision_spec"`
	BaseFilename     string          `json:"base_filename" yaml:"base_filename"`
	RevisionFilename string          `json:"revision_filename" yaml:"revision_filename"`
	Changes          json.RawMessage `json:"changes" yaml:"changes"`
	Mode             string          `json:"mode" yaml:"mode"`
}

// uploadAndOpen runs at the end of `oasdiff changelog --open` (and
// `breaking --open`): it bundles the two specs and the computed changelog,
// encrypts the bundle with a fresh key, uploads only the ciphertext to the
// configured server (oasdiff.com by default; see oasdiffSiteURL), prints the
// resulting review URL (with the key in its #fragment) to stderr, and opens it
// in the default browser. The terminal changelog/breaking output has already
// been printed by the caller; --open is purely additive.
//
// The review URL and browser-fallback notice go to stderr, never stdout, so
// they can't corrupt piped machine-readable output (e.g. changelog
// --format json --open > out.json). stderr is passed in by the caller.
//
// There is no sign-in: the upload is zero-knowledge, so the server stores an
// opaque blob it cannot read and never needs to know who the visitor is. The
// decryption key lives only in the URL fragment on the visitor's machine.
func uploadAndOpen(flags *Flags, stderr io.Writer, isBreaking bool, errs checker.Changes, specInfoPair *load.SpecInfoPair, diffEmpty bool) error {
	// Composed mode (-c) is rejected up front in argument validation
	// (checkOpenWithComposed: --open compares exactly two specs), so it never
	// reaches here.
	baseBytes, baseName, err := readSpecSource(flags.getBase())
	if err != nil {
		return fmt.Errorf("read base spec: %w", err)
	}
	revBytes, revName, err := readSpecSource(flags.getRevision())
	if err != nil {
		return fmt.Errorf("read revision spec: %w", err)
	}

	changesJSON, err := renderChangelogJSON(flags, errs, specInfoPair, isBreaking, diffEmpty)
	if err != nil {
		return fmt.Errorf("render changelog: %w", err)
	}

	mode := "changelog"
	if isBreaking {
		mode = "breaking"
	}

	plaintext, err := json.Marshal(reviewPayload{
		BaseSpec:         string(baseBytes),
		RevisionSpec:     string(revBytes),
		BaseFilename:     baseName,
		RevisionFilename: revName,
		Changes:          changesJSON,
		Mode:             mode,
	})
	if err != nil {
		return fmt.Errorf("marshal review payload: %w", err)
	}

	blob, key, err := encryptReviewPayload(plaintext)
	if err != nil {
		return fmt.Errorf("encrypt review: %w", err)
	}

	// A token switches --open to the authenticated upload; the bundle and key
	// are the same as the free path.
	if token := flags.getReviewToken(); token != "" {
		return uploadAuthenticatedReview(token, flags.getReviewMeta(), blob, key, errs, stderr)
	}

	reviewID, expiresAt, err := postEncryptedReview(blob)
	if err != nil {
		return err
	}

	// The key rides in the URL #fragment. Browsers never transmit the
	// fragment to a server (not in the request path, query, or Referer), so
	// neither the server nor any intermediary sees the key -- only code
	// running in the visitor's own browser can read it.
	reviewURL := fmt.Sprintf("%s/review/e/%s#k=%s",
		oasdiffSiteURL(),
		url.PathEscape(reviewID),
		base64.RawURLEncoding.EncodeToString(key),
	)

	fmt.Fprintf(stderr, "\nOpening %s (expires %s)\n", reviewURL, expiresAt.Format("2006-01-02 15:04 MST"))
	if err := openBrowser(reviewURL); err != nil {
		fmt.Fprintf(stderr, "Could not open browser automatically: %v\nOpen this URL manually: %s\n", err, reviewURL)
	}
	return nil
}

// renderChangelogJSON produces the JSON changelog bytes embedded in the
// encrypted payload. It mirrors the service's /public/changelog rendering
// (FormatJSON + WrapInObject) so the review page consumes identical bytes
// whether the review came through the plaintext path or the encrypted one.
// Color is forced off: the output is data, not a terminal render.
func renderChangelogJSON(flags *Flags, errs checker.Changes, specInfoPair *load.SpecInfoPair, isBreaking, diffEmpty bool) ([]byte, error) {
	formatter, err := formatters.Lookup(string(formatters.FormatJSON), formatters.FormatterOpts{Language: flags.getLang()})
	if err != nil {
		return nil, err
	}
	return formatter.RenderChangelog(
		errs,
		formatters.RenderOpts{WrapInObject: true, ColorMode: checker.ColorNever, IsBreaking: isBreaking, DiffEmpty: diffEmpty},
		specInfoPair.GetBaseVersion(),
		specInfoPair.GetRevisionVersion(),
	)
}

// encryptReviewPayload encrypts plaintext with a freshly generated 256-bit key
// using AES-256-GCM. It returns the upload blob and the key. The blob layout
// is: version(1) || nonce(12) || ciphertext+tag. The key is returned to the
// caller (it goes in the URL fragment) and is never uploaded.
func encryptReviewPayload(plaintext []byte) (blob, key []byte, err error) {
	key = make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	blob = make([]byte, 0, 1+len(nonce)+len(ciphertext))
	blob = append(blob, encryptedReviewBlobVersion)
	blob = append(blob, nonce...)
	blob = append(blob, ciphertext...)
	return blob, key, nil
}

// postEncryptedReview uploads the opaque ciphertext blob to the configured
// server (see oasdiffSiteURL) and returns the assigned review id plus its TTL
// expiry. The request is anonymous (no credentials): the server stores a blob
// it cannot read, so it has nothing to attribute to a user. The body is the
// raw blob; the response is JSON {review_id, expires_at}.
func postEncryptedReview(blob []byte) (string, time.Time, error) {
	endpoint := oasdiffSiteURL() + "/api/encrypted-review"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(blob))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "oasdiff-cli")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("upload to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		ReviewID  string `json:"review_id" yaml:"review_id"`
		ExpiresAt int64  `json:"expires_at" yaml:"expires_at"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", time.Time{}, fmt.Errorf("parse response: %w", err)
	}
	if parsed.ReviewID == "" {
		return "", time.Time{}, fmt.Errorf("response missing review_id: %s", string(respBody))
	}
	return parsed.ReviewID, time.Unix(parsed.ExpiresAt, 0).UTC(), nil
}

// parseReviewMeta splits each "key=value" entry on the first '=' into an opaque
// map (the CLI assigns no meaning to any key). It rejects entries that would
// silently lose or override caller intent rather than swallowing them; see
// TestParseReviewMeta for the exact cases.
func parseReviewMeta(entries []string) (map[string]string, error) {
	meta := make(map[string]string, len(entries))
	for _, e := range entries {
		i := strings.IndexByte(e, '=')
		if i <= 0 {
			return nil, fmt.Errorf("invalid --review-meta entry %q: expected key=value", e)
		}
		key := e[:i]
		if _, dup := meta[key]; dup {
			return nil, fmt.Errorf("duplicate --review-meta key %q", key)
		}
		meta[key] = e[i+1:]
	}
	return meta, nil
}

// reviewChange is one manifest entry sent alongside the encrypted bundle: a
// change's fingerprint (see formatters.ComputeFingerprint) and its level.
type reviewChange struct {
	Fingerprint string `json:"fingerprint" yaml:"fingerprint"`
	Level       int    `json:"level" yaml:"level"`
}

// reviewManifest builds the {fingerprint, level} manifest. Fingerprints are
// computed as the JSON formatter does, so they match the ones in the encrypted
// changelog the page renders.
func reviewManifest(errs checker.Changes) []reviewChange {
	manifest := make([]reviewChange, 0, len(errs))
	for _, change := range errs {
		manifest = append(manifest, reviewChange{
			Fingerprint: formatters.ComputeFingerprint(change.GetId(), change.GetOperation(), change.GetPath(), change.GetArgs()),
			Level:       int(change.GetLevel()),
		})
	}
	return manifest
}

// uploadAuthenticatedReview posts the encrypted bundle to the token endpoint and
// prints the returned review URL (key in its #fragment) plus the status. Errors
// are returned for the caller to demote to a warning, like the free path.
func uploadAuthenticatedReview(token string, metaEntries []string, blob, key []byte, errs checker.Changes, stderr io.Writer) error {
	metadata, err := parseReviewMeta(metaEntries)
	if err != nil {
		return err
	}

	body, err := json.Marshal(struct {
		Ciphertext []byte            `json:"ciphertext" yaml:"ciphertext"`
		Metadata   map[string]string `json:"metadata" yaml:"metadata"`
		Changes    []reviewChange    `json:"changes" yaml:"changes"`
	}{
		Ciphertext: blob,
		Metadata:   metadata,
		Changes:    reviewManifest(errs),
	})
	if err != nil {
		return fmt.Errorf("marshal authenticated review: %w", err)
	}

	endpoint := fmt.Sprintf("%s/tenants/%s/encrypted-pr-review", oasdiffAPIBaseURL(), url.PathEscape(token))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "oasdiff-cli")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		ReviewToken string `json:"review_token" yaml:"review_token"`
		ReviewURL   string `json:"review_url" yaml:"review_url"`
		Gate        struct {
			State            string `json:"state" yaml:"state"`
			BreakingTotal    int    `json:"breaking_total" yaml:"breaking_total"`
			BreakingApproved int    `json:"breaking_approved" yaml:"breaking_approved"`
		} `json:"gate" yaml:"gate"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if parsed.ReviewURL == "" {
		return fmt.Errorf("response missing review_url: %s", string(respBody))
	}

	// The key rides in the URL #fragment exactly as the free path encodes it:
	// browsers never transmit the fragment, so the server stores ciphertext it
	// cannot read.
	reviewURL := parsed.ReviewURL + "#k=" + base64.RawURLEncoding.EncodeToString(key)

	fmt.Fprintf(stderr, "\nOpening %s\n", reviewURL)
	// The status is printed verbatim, on its own grep-friendly line. The CLI
	// does not interpret or branch on it; the caller acts on it.
	fmt.Fprintf(stderr, "oasdiff: review status: %s\n", parsed.Gate.State)
	if err := openBrowser(reviewURL); err != nil {
		fmt.Fprintf(stderr, "Could not open browser automatically: %v\nOpen this URL manually: %s\n", err, reviewURL)
	}
	return nil
}

// readSpecSource returns the raw bytes of a spec source and a display
// filename for the upload. --open supports file and git-ref sources (the
// git-ref read, including blob-hash handling, lives in the load package);
// stdin and URL sources are rejected here because the upload requires bytes
// the CLI can attribute to a filename.
func readSpecSource(source *load.Source) ([]byte, string, error) {
	if source == nil {
		return nil, "", errors.New("spec source is required")
	}
	if source.IsStdin() {
		return nil, "", errors.New("--open does not support stdin (use a file path or git ref)")
	}
	if !source.IsFile() && !source.IsGitRevision() {
		return nil, "", fmt.Errorf("--open does not support source type for %q", source.Path)
	}
	body, err := source.ReadRaw()
	if err != nil {
		return nil, "", err
	}
	// DisplayPath strips the "<ref>:" prefix for git sources; Base trims any
	// directory so the upload's filename is just "openapi.yaml".
	return body, filepath.Base(source.DisplayPath()), nil
}

// openBrowser opens the URL in the default browser. xdg-open / open / start
// cover Linux, macOS, and Windows. Non-zero exit from the opener is treated
// as a soft failure — the caller prints the URL for the user to follow
// manually. Notable absence: a CI / headless detection. CI users wouldn't
// run --open in the first place; if they do, they get a non-fatal error
// and the printed URL.
func openBrowser(targetURL string) error {
	if _, err := url.Parse(targetURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", targetURL)
	default:
		return fmt.Errorf("don't know how to open a browser on %s", runtime.GOOS)
	}
	return cmd.Start()
}
