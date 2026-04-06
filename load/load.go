package load

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
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
// Relative $refs (e.g. "./schemas/pet.yaml") are resolved by kin-openapi into paths like
// "origin/main:multi-file/schemas/pet.yaml". We install a ReadFromURIFunc so the loader
// recognises that pattern and reads referenced files via "git show" too.
func loadFromGitRevision(loader *openapi3.Loader, gitRef string) (*openapi3.T, error) {
	out, err := gitShow(gitRef)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec from git revision %q: %w", gitRef, err)
	}

	// Allow external refs: they are resolved via "git show" (no network), so this is safe.
	loader.IsExternalRefsAllowed = true

	// Install a ReadFromURIFunc so relative $refs inside a git-revision spec are also
	// fetched via "git show" rather than opened as literal file paths.
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
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
	return loader.LoadFromDataWithPath(out, u)
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
