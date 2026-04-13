package load

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/oasdiff/kin-openapi/openapi3"
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
func loadFromGitRevision(loader *openapi3.Loader, gitRef string) (*openapi3.T, error) {
	out, err := gitShow(gitRef)
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

	// kin-openapi calls ReadFromURIFunc before checking IsExternalRefsAllowed, so the
	// caller's --allow-external-refs setting is still enforced for non-git refs via
	// DefaultReadFromURI.
	loaderCopy.ReadFromURIFunc = func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		p := filepath.FromSlash(location.Path)
		if isGitRevision(p) {
			return gitShow(p)
		}
		return openapi3.DefaultReadFromURI(loader, location)
	}

	// Use the full gitRef as the URL path so each revision gets a unique cache key in the
	// loader's visitedDocuments map (e.g. "origin/main:openapi.yaml" vs "HEAD:openapi.yaml").
	// Using only the file portion would cause both refs to share the key "openapi.yaml" and
	// the loader would return the cached base spec for the revision.
	u := &url.URL{Path: filepath.ToSlash(gitRef)}
	return loaderCopy.LoadFromDataWithPath(out, u)
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
