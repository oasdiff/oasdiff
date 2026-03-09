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

// NewSource creates a Source by categorizing the input path as stdin, URL, or file.
// This function is intentionally infallible (does not return an error) to allow
// clean usage in struct literal initialization and avoid error handling boilerplate
// in hundreds of call sites throughout the codebase.
//
// Design rationale:
//   - Valid http/https URLs are categorized as SourceTypeURL
//   - "-" is categorized as SourceTypeStdin
//   - Everything else (including URLs with invalid schemes like ftp://) is categorized as SourceTypeFile
//   - Actual validation and error handling occurs later when the Loader interface methods
//     (LoadFromURI, LoadFromFile, LoadFromStdin) are called
//
// This design provides clean separation of concerns: NewSource categorizes inputs,
// while Loader implementations handle validation and produce appropriate error messages.
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
