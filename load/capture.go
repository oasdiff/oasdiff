package load

import (
	"net/url"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// sourceCapture holds the raw bytes of every file the loader reads, the root
// spec and each $ref'd file, keyed by location.String() -- the exact value
// kin-openapi records as origin.Key.File -- so a file is found by the File on
// any element's origin with a direct lookup, no normalization.
type sourceCapture struct {
	files map[string][]byte
}

func newSourceCapture() *sourceCapture {
	return &sourceCapture{files: map[string][]byte{}}
}

func (c *sourceCapture) record(key string, data []byte) {
	if c == nil {
		return
	}
	if _, seen := c.files[key]; !seen {
		c.files[key] = append([]byte(nil), data...)
	}
}

// asStrings returns the captured path -> content map, so a consumer can slice an
// element by its origin File.
func (c *sourceCapture) asStrings() map[string]string {
	out := make(map[string]string, len(c.files))
	for k, v := range c.files {
		out[k] = string(v)
	}
	return out
}

// recordingReader wraps a ReadFromURIFunc (or kin's default when inner is nil)
// to record every successfully read file into the capture, keyed by
// location.String() to match origin.Key.File (kin derives it the same way, see
// its marsh.go), so a captured file is found by an element's origin File on
// every platform. It is installed only when the load is given a capture;
// ordinary loads install no recorder and pay nothing.
func recordingReader(inner openapi3.ReadFromURIFunc, capture *sourceCapture, root string) openapi3.ReadFromURIFunc {
	return func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		if err := allowLocation(loader, location, root); err != nil {
			return nil, err
		}
		var (
			data []byte
			err  error
		)
		if inner != nil {
			data, err = inner(loader, location)
		} else {
			data, err = openapi3.DefaultReadFromURI(loader, location)
		}
		if err == nil {
			capture.record(location.String(), data)
		}
		return data, err
	}
}

// allowLocation applies the loader's external-refs policy to a location a custom
// reader is about to read.
//
// kin-openapi applies that policy itself only when no custom ReadFromURIFunc is
// installed; with one it hands the policy to the reader (see
// Loader.loadSingleElementFromURI). A custom reader that does not apply it
// therefore turns --allow-external-refs=false into a no-op, which is how
// GHSA-2jcc-mxv7-p3f9 happened on the git-revision path and how it happened
// again here. Both readers now go through this.
//
// With external refs disallowed a spec must be self-contained, so only the
// document the load was asked for may be read. That matches what kin-openapi
// refuses on its own: any $ref to another document, a sibling file included.
func allowLocation(loader *openapi3.Loader, location *url.URL, root string) error {
	if loader == nil || loader.IsExternalRefsAllowed || root == "" {
		return nil
	}
	if sameLocation(location, root) {
		return nil
	}
	return &ExternalRefError{Ref: location.String()}
}

// sameLocation reports whether location is the root document. kin-openapi may
// hand the reader an absolute or a cleaned form of the path it was given, so
// compare both as given and resolved.
func sameLocation(location *url.URL, root string) bool {
	if location.String() == root {
		return true
	}
	lp := filepath.FromSlash(location.Path)
	if lp == root || filepath.Clean(lp) == filepath.Clean(root) {
		return true
	}
	a, errA := filepath.Abs(lp)
	b, errB := filepath.Abs(root)
	return errA == nil && errB == nil && a == b
}
