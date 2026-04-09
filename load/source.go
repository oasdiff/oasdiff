package load

import (
	"fmt"
	"net/url"
	"strings"
)

type SourceType int

const (
	SourceTypeStdin SourceType = iota
	SourceTypeURL
	SourceTypeFile
	SourceTypeGitRevision
)

type Source struct {
	Path string
	Uri  *url.URL
	Type SourceType
}

// NewSource creates a Source by categorizing the input path as stdin, URL, git revision, or file.
// This function is intentionally infallible (does not return an error) to allow
// clean usage in struct literal initialization and avoid error handling boilerplate
// in hundreds of call sites throughout the codebase.
//
// Categorization rules (evaluated in order):
//   - "-" → SourceTypeStdin
//   - Git revision syntax (e.g. "HEAD:openapi.yaml", "origin/main:api/openapi.yaml") → SourceTypeGitRevision
//   - Valid http/https URLs → SourceTypeURL
//   - Everything else (including URLs with unsupported schemes) → SourceTypeFile
//
// Git revision syntax is "<ref>:<path>" where <ref> is any git ref (branch, tag, commit SHA,
// or expressions like HEAD~1) and <path> is the file path within the repo. Multi-file specs
// with relative $refs are fully supported — referenced files are also read via "git show".
//
// Actual validation and error handling occurs later when the source is loaded, providing
// clean separation of concerns between categorization and I/O.
func NewSource(path string) *Source {
	if path == "-" {
		return &Source{
			Path: "stdin",
			Type: SourceTypeStdin,
		}
	}

	if isGitRevision(path) {
		return &Source{
			Path: path,
			Type: SourceTypeGitRevision,
		}
	}

	uri, err := getURL(path)
	if err == nil {
		return &Source{
			Path: path,
			Type: SourceTypeURL,
			Uri:  uri,
		}
	}

	return &Source{
		Path: path,
		Type: SourceTypeFile,
	}
}

// isGitRevision returns true if path looks like a git revision spec (e.g. "origin/main:openapi.yaml" or "HEAD~1:spec.yaml").
// It excludes URLs (contain "://") and Windows drive letters (single uppercase letter before ":").
func isGitRevision(path string) bool {
	idx := strings.Index(path, ":")
	if idx < 0 || strings.Contains(path, "://") {
		return false
	}
	// Single uppercase letter before colon = Windows drive letter (C:, D:, …)
	if idx == 1 && path[0] >= 'A' && path[0] <= 'Z' {
		return false
	}
	return true
}

func (source *Source) String() string {
	return source.Path
}

func (source *Source) Out() string {
	if source.IsStdin() {
		return source.Path
	}
	return fmt.Sprintf("%q", source.Path)
}

func (source *Source) IsStdin() bool {
	return source.Type == SourceTypeStdin
}

func (source *Source) IsFile() bool {
	return source.Type == SourceTypeFile
}

func (source *Source) IsGitRevision() bool {
	return source.Type == SourceTypeGitRevision
}

// DisplayPath returns the path suitable for display and source-location reporting.
// For git revisions it strips the ref prefix (e.g. "origin/main:openapi.yaml" → "openapi.yaml").
func (source *Source) DisplayPath() string {
	if source.Type != SourceTypeGitRevision {
		return source.Path
	}
	return source.Path[strings.Index(source.Path, ":")+1:]
}
