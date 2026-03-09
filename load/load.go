package load

import (
	"errors"
	"fmt"
	"net/url"
	"os"
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
// It runs "git show <ref>" to obtain the content, writes it to a temp file in the same directory as
// the spec path (so that relative $refs resolve correctly), and loads from that temp file.
func loadFromGitRevision(loader Loader, gitRef string) (*openapi3.T, error) {
	out, err := exec.Command("git", "show", gitRef).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("failed to load spec from git revision %q: %s", gitRef, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("failed to load spec from git revision %q: %w", gitRef, err)
	}

	specPath := gitRef[strings.Index(gitRef, ":")+1:]
	tmpFile, err := os.CreateTemp(filepath.Dir(specPath), "oasdiff-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for git revision %q: %w", gitRef, err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(out); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file for git revision %q: %w", gitRef, err)
	}
	tmpFile.Close()

	return loader.LoadFromFile(tmpFile.Name())
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
