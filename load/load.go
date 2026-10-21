package load

import (
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// from is a convenience function that opens an OpenAPI spec from a URL or a
// local path based on the format of the path parameter. When capture is
// non-nil, every branch except stdin records the files the loader reads (the
// root and each $ref'd file) into it; stdin ignores it, since the review
// bundle (the capture's only consumer) does not accept stdin sources.
func from(loader *openapi3.Loader, source *Source, capture *sourceCapture) (*openapi3.T, error) {
	switch source.Type {
	case SourceTypeStdin:
		return loader.LoadFromStdin()
	case SourceTypeURL:
		return withCapture(loader, capture, source.Uri.String()).LoadFromURI(source.Uri)
	case SourceTypeGitRevision:
		return loadFromGitRevision(loader, source.Path, source.Fetch, capture)
	default:
		return withCapture(loader, capture, source.Path).LoadFromFile(source.Path)
	}
}

// withCapture records the files a load reads and refuses any $ref to a document
// other than rootDocument while external refs are disallowed.
func withCapture(loader *openapi3.Loader, capture *sourceCapture, rootDocument string) *openapi3.Loader {
	return withRecordingReader(loader, capture, refPolicy{root: rootDocument})
}

// withCaptureInnerEnforcesPolicy is withCapture for a loader whose own
// ReadFromURIFunc already applies the external-refs policy, so the recorder
// must not apply it a second time against a root path it does not have.
func withCaptureInnerEnforcesPolicy(loader *openapi3.Loader, capture *sourceCapture) *openapi3.Loader {
	return withRecordingReader(loader, capture, refPolicy{innerEnforces: true})
}

// withRecordingReader returns loader unchanged when capture is nil, otherwise a
// shallow copy whose ReadFromURIFunc records every file the loader reads and
// applies policy to each location before reading it. The copy keeps the recorder
// off the caller's loader while still sharing its visitedDocuments cache (the
// same reason loadFromGitRevision copies the loader).
func withRecordingReader(loader *openapi3.Loader, capture *sourceCapture, policy refPolicy) *openapi3.Loader {
	if capture == nil {
		return loader
	}
	lc := *loader
	lc.ReadFromURIFunc = recordingReader(lc.ReadFromURIFunc, capture, policy)
	return &lc
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
