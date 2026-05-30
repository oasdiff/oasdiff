package load

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// from is a convenience function that opens an OpenAPI spec from a URL or a local path based on the format of the path parameter
func from(loader *openapi3.Loader, source *Source) (*openapi3.T, error) {

	switch source.Type {
	case SourceTypeStdin:
		return loader.LoadFromStdin()
	case SourceTypeURL:
		return loader.LoadFromURI(source.Uri)
	case SourceTypeGitRevision:
		return loadFromGitRevision(loader, source.Path)
	default:
		return loader.LoadFromFile(source.Path)
	}
}

// loadFromGitRevision loads an OpenAPI spec from a git revision reference (e.g. "origin/main:openapi.yaml").
// It runs "git show <ref>" to obtain the content and loads it via LoadFromDataWithPath so that
// relative $refs are resolved against the spec's path.
//
// Relative $refs (e.g. "./schemas/pet.yaml") are resolved by kin-openapi; we install a
// JoinFunc so the "<rev>:" prefix is preserved (the default path.Dir strips it), and a
// ReadFromURIFunc so referenced files are read via "git show" too.
//
// Blob-hash refs (e.g. "abc1234:openapi.yaml" where abc1234 is a git blob object hash)
// are supported via readGitRefContent. The path portion is preserved on the ref string
// passed downstream for source labels and relative $ref resolution.
func loadFromGitRevision(loader *openapi3.Loader, gitRef string) (*openapi3.T, error) {
	out, err := readGitRefContent(gitRef)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from git revision %q: %w", gitRef, err)
	}

	// Copy the loader so we can install custom resolvers without mutating the caller's
	// instance. The shallow copy shares the unexported visitedDocuments cache, which is
	// intentional: common $ref files are not fetched twice, and the unique gitRef-based
	// URL keys prevent collisions between base and revision entries.
	loaderCopy := *loader

	// JoinFunc preserves the "<rev>:" prefix when resolving relative $refs.
	// Without this, path.Dir("origin/main:openapi.yaml") returns "origin" and
	// "./schemas/pet.yaml" resolves to "origin/schemas/pet.yaml" instead of
	// "origin/main:schemas/pet.yaml".
	loaderCopy.JoinFunc = func(basePath, relativePath *url.URL) *url.URL {
		if basePath == nil {
			return relativePath
		}
		result := *basePath
		base := basePath.Path
		if i := strings.IndexByte(base, ':'); i >= 0 {
			result.Path = base[:i+1] + path.Join(path.Dir(base[i+1:]), relativePath.Path)
		} else {
			result.Path = path.Join(path.Dir(base), relativePath.Path)
		}
		return &result
	}

	// kin-openapi skips its own allowsExternalRefs policy check whenever a custom
	// ReadFromURIFunc is installed (see openapi3.Loader.resolveComponent), so we
	// must enforce IsExternalRefsAllowed here. Intra-git relative $refs keep their
	// "<rev>:" prefix (via JoinFunc above) and are read with "git show". Any other
	// $ref is genuinely external — an http(s) URL or a local file path outside the
	// git tree — so honor IsExternalRefsAllowed: with --allow-external-refs=false a
	// spec loaded from an (attacker-controlled) git blob must not be able to read
	// local files or reach arbitrary URLs (SSRF). This matches the non-git load
	// path, where kin-openapi enforces the same policy itself.
	// blockedRef records the first external $ref we refuse, so loadFromGitRevision
	// can return a typed *ExternalRefError regardless of how kin-openapi wraps the
	// error from ReadFromURIFunc — callers map the type to a dedicated exit code.
	var blockedRef string
	loaderCopy.ReadFromURIFunc = func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		p := filepath.FromSlash(location.Path)
		if isGitRevision(p) {
			return gitShow(p)
		}
		if !loader.IsExternalRefsAllowed {
			if blockedRef == "" {
				blockedRef = location.String()
			}
			return nil, &ExternalRefError{Ref: location.String()}
		}
		return openapi3.DefaultReadFromURI(loader, location)
	}

	// Use the full gitRef as the URL path so each revision gets a unique cache key in the
	// loader's visitedDocuments map (e.g. "origin/main:openapi.yaml" vs "HEAD:openapi.yaml").
	// Using only the file portion would cause both refs to share the key "openapi.yaml" and
	// the loader would return the cached base spec for the revision.
	u := &url.URL{Path: filepath.ToSlash(gitRef)}
	t, err := loaderCopy.LoadFromDataWithPath(out, u)
	if err != nil && blockedRef != "" {
		// Return the typed error even if kin-openapi wrapped ours in plain text,
		// so callers can errors.As it to a dedicated exit code.
		return nil, &ExternalRefError{Ref: blockedRef}
	}
	return t, err
}

// readGitRefContent fetches the OpenAPI spec bytes for a git ref. For the
// `<commit-or-tag>:<path>` form it runs `git show <ref>:<path>`. For the
// `<blob-hex>:<path>` form (where the part before `:` is a blob object hash
// rather than a commit/tree/tag) it runs `git show <blob-hex>` directly, since
// `git show <blob-hex>:<path>` is not valid git syntax. Callers keep the full
// `<ref>:<path>` string for the loader's URL key so source labels and relative
// $ref resolution stay correct.
func readGitRefContent(gitRef string) ([]byte, error) {
	if hex, ok := blobHashFromRef(gitRef); ok {
		return gitShow(hex)
	}
	return gitShow(gitRef)
}

// blobHashFromRef returns (hex, true) when the part before `:` in gitRef names
// a git blob object (a stored file content, not a commit/tree/tag). Returns
// ("", false) when the ref isn't a blob, when git cannot resolve it, or when
// the input has no `:` separator. The cost is one extra `git cat-file -t` call
// per spec load on the git path, which git serves in single-digit milliseconds.
func blobHashFromRef(gitRef string) (string, bool) {
	idx := strings.Index(gitRef, ":")
	if idx < 0 {
		return "", false
	}
	ref := gitRef[:idx]
	out, err := exec.Command("git", "cat-file", "-t", ref).Output()
	if err != nil {
		return "", false
	}
	if strings.TrimSpace(string(out)) != "blob" {
		return "", false
	}
	return ref, true
}

// gitShow runs "git show <ref>" and returns its stdout, or a descriptive error.
func gitShow(ref string) ([]byte, error) {
	out, err := exec.Command("git", "show", ref).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("is git installed and in PATH?: %w", err)
	}
	return out, nil
}

func getURL(rawURL string) (*url.URL, error) {
	url, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, err
	}

	if !isValidScheme(url.Scheme) {
		return nil, fmt.Errorf("invalid scheme: %s", url.Scheme)
	}

	return url, nil
}

func isValidScheme(scheme string) bool {

	switch scheme {
	case "http":
	case "https":
	default:
		return false
	}

	return true
}
