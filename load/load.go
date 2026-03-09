package load

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
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
func loadFromGitRevision(loader *openapi3.Loader, gitRef string) (*openapi3.T, error) {
	out, err := exec.Command("git", "show", gitRef).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("failed to load spec from git revision %q: %s", gitRef, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("failed to load spec from git revision %q (is git installed and in PATH?): %w", gitRef, err)
	}

	// Use the full gitRef as the URL path so each revision gets a unique cache key in the
	// loader's visitedDocuments map (e.g. "origin/main:openapi.yaml" vs "HEAD:openapi.yaml").
	// Using only the file portion would cause both refs to share the key "openapi.yaml" and
	// the loader would return the cached base spec for the revision.
	u := &url.URL{Path: filepath.ToSlash(gitRef)}
	return loader.LoadFromDataWithPath(out, u)
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
