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
		return loadFromGitRevision(loader, source.Path, source.Fetch)
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
func loadFromGitRevision(loader *openapi3.Loader, gitRef string, fetch bool) (*openapi3.T, error) {
	out, err := readGitRefContent(gitRef, fetch)
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
func readGitRefContent(gitRef string, fetch bool) ([]byte, error) {
	if hex, ok := blobHashFromRef(gitRef); ok {
		return gitShow(hex)
	}
	out, err := gitShow(gitRef)
	if err == nil {
		return out, nil
	}
	// "git show" failed. When --fetch is set and the only problem is that the
	// commit isn't in this clone, fetch that commit from origin and retry once.
	// This is the opt-in counterpart to hintMissingObject: instead of telling
	// the reviewer to run "git fetch origin <ref>" themselves, oasdiff runs it
	// for them (mutating the repo by downloading objects). A path that doesn't
	// exist within an already-local commit is not a fetch problem, so we only
	// fetch when the commit itself is absent.
	if fetch {
		ref := refBeforeColon(gitRef)
		if !commitExistsLocally(ref) {
			if ferr := gitFetch(ref); ferr != nil {
				return nil, fmt.Errorf(
					"failed to fetch git revision %q from origin: %w\n\noriginal error: %v",
					ref, ferr, err)
			}
			if out, err = gitShow(gitRef); err == nil {
				return out, nil
			}
		}
	}
	return nil, hintMissingObject(gitRef, err)
}

// hintMissingObject augments a "git show" failure with a recovery hint when the
// failure is that the commit named before the colon is not present in the local
// clone. By default oasdiff resolves "<ref>:<path>" against local objects only
// and does not fetch or otherwise mutate the repository, so a reviewer who
// hasn't fetched the PR branch (or a shallow clone lacking the base commit)
// would otherwise be left with git's terse error. We hand them the exact command
// to fetch the commit themselves, and point at the --fetch flag that does it
// automatically; the repository stays untouched unless they opt in.
//
// The hint fires only when the commit genuinely doesn't resolve locally. A path
// that doesn't exist within an existing commit is returned unchanged (its message
// is already actionable), and so is a "git not installed" failure (there is
// nothing to fetch). git's own "git show" message conflates the missing-commit
// and missing-path cases when the path happens to exist in the working tree, so
// we probe the object directly rather than parse the message text.
func hintMissingObject(gitRef string, err error) error {
	if strings.Contains(err.Error(), "is git installed") {
		return err
	}
	ref := refBeforeColon(gitRef)
	if commitExistsLocally(ref) {
		return err
	}
	return fmt.Errorf("%w\n\n"+
		"oasdiff reads git revisions from local objects only and does not modify your repository.\n"+
		"Commit %q is not in this clone. Fetch it yourself with:\n\n"+
		"    git fetch origin %s\n\n"+
		"then re-run oasdiff, or re-run with --fetch to let oasdiff fetch it for you", err, ref, ref)
}

// refBeforeColon returns the "<ref>" portion of a "<ref>:<path>" git revision,
// or the whole string when there is no colon.
func refBeforeColon(gitRef string) string {
	if before, _, found := strings.Cut(gitRef, ":"); found {
		return before
	}
	return gitRef
}

// gitFetch downloads the named commit/ref from the "origin" remote into the
// local object store so a subsequent "git show <ref>:<path>" can resolve it. It
// mutates the repository (adds objects; no local ref or branch is moved), which
// is why it runs only under the opt-in --fetch flag. It mirrors the manual
// "git fetch origin <ref>" that hintMissingObject suggests when --fetch is off.
func gitFetch(ref string) error {
	if out, err := exec.Command("git", "fetch", "origin", ref).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// commitExistsLocally reports whether ref names an object already present in the
// local repository, via "git cat-file -e". Returns false when the object is
// absent or when git cannot run.
func commitExistsLocally(ref string) bool {
	return exec.Command("git", "cat-file", "-e", ref).Run() == nil
}

// blobHashFromRef returns (hex, true) when the part before `:` in gitRef names
// a git blob object (a stored file content, not a commit/tree/tag). Returns
// ("", false) when the ref isn't a blob, when git cannot resolve it, or when
// the input has no `:` separator. The cost is one extra `git cat-file -t` call
// per spec load on the git path, which git serves in single-digit milliseconds.
func blobHashFromRef(gitRef string) (string, bool) {
	before, _, ok := strings.Cut(gitRef, ":")
	if !ok {
		return "", false
	}
	ref := before
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
