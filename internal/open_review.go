package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/oasdiff/oasdiff/load"
)

// oasdiffSiteURL is the base URL of the web product. Overridable via the
// OASDIFF_URL env var for local development and tests; in production this
// is the canonical oasdiff.com.
func oasdiffSiteURL() string {
	if u := os.Getenv("OASDIFF_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://www.oasdiff.com"
}

// uploadAndOpen runs at the end of `oasdiff changelog --open` (and
// `breaking --open`): uploads the two spec sources to oasdiff.com, prints
// the resulting review URL, and opens it in the default browser. The
// terminal changelog/breaking output has already been printed by the caller;
// --open is purely additive.
//
// Sign-in: if no access token is stored locally, the CLI runs the first-run
// browser handshake. See enterprise/docs/cli-local-review.md for the design.
func uploadAndOpen(flags *Flags, stdout io.Writer) error {
	baseBytes, baseName, err := readSpecSource(flags.getBase())
	if err != nil {
		return fmt.Errorf("read base spec: %w", err)
	}
	revBytes, revName, err := readSpecSource(flags.getRevision())
	if err != nil {
		return fmt.Errorf("read revision spec: %w", err)
	}

	options := buildSemanticOptionsJSON(flags)

	accessToken, err := readOrMintAccessToken(stdout)
	if err != nil {
		return err
	}

	reviewURL, expiresAt, err := postPreviewReview(accessToken, baseName, baseBytes, revName, revBytes, options)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "\nOpening %s (expires %s)\n", reviewURL, expiresAt.Format("2006-01-02 15:04 MST"))
	if err := openBrowser(reviewURL); err != nil {
		fmt.Fprintf(stdout, "Could not open browser automatically: %v\nOpen this URL manually: %s\n", err, reviewURL)
	}
	return nil
}

// readSpecSource returns the raw bytes of a spec source and a display
// filename. For files: ReadFile + basename. For git refs: `git show` and
// the filename portion. URL and stdin sources are not supported by --open
// in v1 — the upload requires bytes the CLI can attribute to a filename.
func readSpecSource(source *load.Source) ([]byte, string, error) {
	if source == nil {
		return nil, "", errors.New("spec source is required")
	}
	switch {
	case source.IsStdin():
		return nil, "", errors.New("--open does not support stdin (use a file path or git ref)")
	case source.IsGitRevision():
		raw := source.Path
		// "<ref>:<path>". For blob hashes, `git show <hex>:<path>` is invalid
		// — the path is for display only. NewSpecInfo handles this distinction
		// internally; here we just call git show <full-ref> and let it speak.
		cmd := exec.Command("git", "show", raw)
		out, err := cmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
				return nil, "", fmt.Errorf("git show %s: %s", raw, strings.TrimSpace(string(exitErr.Stderr)))
			}
			// Blob-hash refs require the explicit blob-only call.
			if hex, _, ok := splitRefPath(raw); ok && looksLikeHex(hex) {
				out2, err2 := exec.Command("git", "show", hex).Output()
				if err2 == nil {
					return out2, displayNameFromRef(raw), nil
				}
			}
			return nil, "", fmt.Errorf("git show %s: %w", raw, err)
		}
		return out, displayNameFromRef(raw), nil
	case source.IsFile():
		body, err := os.ReadFile(source.Path)
		if err != nil {
			return nil, "", fmt.Errorf("read %s: %w", source.Path, err)
		}
		return body, filepath.Base(source.Path), nil
	default:
		return nil, "", fmt.Errorf("--open does not support source type for %q", source.Path)
	}
}

func splitRefPath(s string) (string, string, bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", false
	}
	return s[:idx], s[idx+1:], true
}

func looksLikeHex(s string) bool {
	if len(s) < 4 || len(s) > 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func displayNameFromRef(ref string) string {
	if _, path, ok := splitRefPath(ref); ok && path != "" {
		return filepath.Base(path)
	}
	return ref
}

// buildSemanticOptionsJSON returns a JSON blob of the semantic comparison
// flags the visitor passed (flatten-allof, flatten-params, etc.). Filtering
// and presentation flags are deliberately not included — the web UI handles
// those interactively. See enterprise/docs/cli-local-review.md.
func buildSemanticOptionsJSON(flags *Flags) string {
	opts := map[string]any{}
	if flags.getFlattenAllOf() {
		opts["flatten-allof"] = true
	}
	if flags.getFlattenParams() {
		opts["flatten-params"] = true
	}
	if flags.getCaseInsensitiveHeaders() {
		opts["case-insensitive-headers"] = true
	}
	if flags.getAutoUpgrade() {
		opts["auto-upgrade"] = true
	}
	v := flags.getViper()
	if v.GetBool("match-inline-refs") {
		opts["match-inline-refs"] = true
	}
	if v.GetBool("include-path-params") {
		opts["include-path-params"] = true
	}
	if len(opts) == 0 {
		return ""
	}
	b, _ := json.Marshal(opts)
	return string(b)
}

// postPreviewReview uploads the two spec files to oasdiff.com and returns
// the rendered review URL plus its TTL expiry timestamp.
func postPreviewReview(accessToken, baseName string, baseBytes []byte, revName string, revBytes []byte, options string) (string, time.Time, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := addFilePart(writer, "base", baseName, baseBytes); err != nil {
		return "", time.Time{}, err
	}
	if err := addFilePart(writer, "revision", revName, revBytes); err != nil {
		return "", time.Time{}, err
	}
	if options != "" {
		if err := writer.WriteField("options", options); err != nil {
			return "", time.Time{}, fmt.Errorf("write options field: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return "", time.Time{}, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint := oasdiffSiteURL() + "/api/preview-review"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, body)
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "oasdiff-cli")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("upload to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		// Stale token. Clear it and ask the user to retry, which will trigger
		// a fresh first-run handshake. We do not retry automatically because
		// running the OAuth dance silently after a successful diff would
		// surprise the user.
		_ = deleteStoredAccessToken()
		return "", time.Time{}, fmt.Errorf("stored credentials are no longer valid; cleared them. Run the command again to sign in")
	}
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		ReviewId  string `json:"review_id"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", time.Time{}, fmt.Errorf("parse response: %w", err)
	}
	if parsed.ReviewId == "" {
		return "", time.Time{}, fmt.Errorf("response missing review_id: %s", string(respBody))
	}
	reviewURL := oasdiffSiteURL() + "/review/local/" + url.PathEscape(parsed.ReviewId)
	expiresAt := time.Unix(parsed.ExpiresAt, 0).UTC()
	return reviewURL, expiresAt, nil
}

func addFilePart(writer *multipart.Writer, fieldName, fileName string, body []byte) error {
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("create form file %s: %w", fieldName, err)
	}
	if _, err := part.Write(body); err != nil {
		return fmt.Errorf("write %s body: %w", fieldName, err)
	}
	return nil
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

// credentialsPath returns the platform-appropriate path to the CLI auth
// credentials file. Linux: $XDG_CONFIG_HOME/oasdiff/credentials. macOS:
// ~/Library/Application Support/oasdiff/credentials. Windows:
// %AppData%\oasdiff\credentials.
func credentialsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(configDir, "oasdiff", "credentials"), nil
}

// readStoredAccessToken loads a previously issued access token. Returns
// ("", nil) when no token is stored (first run). Returns ("", err) for
// real filesystem failures (permission denied, etc.).
func readStoredAccessToken() (string, error) {
	path, err := credentialsPath()
	if err != nil {
		return "", err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("read credentials: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

// writeStoredAccessToken persists a freshly minted access token. File mode
// 0600 because the token is a credential — same care as ~/.ssh/id_*.
func writeStoredAccessToken(token string) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	return os.WriteFile(path, []byte(token+"\n"), 0600)
}

func deleteStoredAccessToken() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// readOrMintAccessToken returns the stored access token, or runs the
// first-run browser sign-in if none is stored. The sign-in flow opens a
// random-port HTTP listener on 127.0.0.1, opens the visitor's browser at
// /cli-login?port=N, waits for the POST callback, captures the token from
// the form body, and writes it to credentialsPath() before returning.
func readOrMintAccessToken(stdout io.Writer) (string, error) {
	token, err := readStoredAccessToken()
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
	fmt.Fprintf(stdout, "\nFirst time on this machine. Opening browser to sign in...\n")
	token, err = signInViaBrowser(stdout)
	if err != nil {
		return "", err
	}
	if err := writeStoredAccessToken(token); err != nil {
		return "", err
	}
	path, _ := credentialsPath()
	fmt.Fprintf(stdout, "Stored credentials at %s\n", path)
	return token, nil
}

// signInViaBrowser runs the localhost-handoff dance. Binds to 127.0.0.1 on a
// random high port; opens the visitor's browser at /cli-login?port=N; waits
// up to 5 minutes for a POST whose body contains the access_token form field;
// returns the token.
//
// Edge cases (SSH / WSL / Codespaces where the browser and CLI run on
// different hosts) are not handled in v1; the listener times out and the
// caller surfaces the failure to the visitor. The deferred v2 fix is a
// copy-paste fallback — see enterprise/docs/cli-local-review.md.
func signInViaBrowser(stdout io.Writer) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("bind local listener: %w", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			errCh <- fmt.Errorf("parse callback form: %w", err)
			return
		}
		token := strings.TrimSpace(r.PostFormValue("access_token"))
		if token == "" {
			http.Error(w, "missing access_token", http.StatusBadRequest)
			errCh <- errors.New("callback missing access_token")
			return
		}
		fmt.Fprintln(w, "<!doctype html><html><body style=\"font-family:sans-serif;text-align:center;padding:4rem\"><h2>oasdiff CLI is signed in</h2><p>You can close this tab and return to your terminal.</p></body></html>")
		tokenCh <- token
	})

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = server.Serve(ln) }()
	defer func() { _ = server.Shutdown(context.Background()) }()

	loginURL := fmt.Sprintf("%s/cli-login?port=%d", oasdiffSiteURL(), port)
	if err := openBrowser(loginURL); err != nil {
		fmt.Fprintf(stdout, "Could not open browser automatically: %v\nOpen this URL manually: %s\n", err, loginURL)
	} else {
		fmt.Fprintf(stdout, "  -> %s\n", loginURL)
	}

	select {
	case token := <-tokenCh:
		return token, nil
	case err := <-errCh:
		return "", err
	case <-time.After(5 * time.Minute):
		return "", errors.New("sign-in timed out after 5 minutes — no callback received from the browser")
	}
}
