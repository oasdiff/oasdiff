package load

import (
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// sourceCapture holds the raw bytes of every file the loader reads, the root
// spec and each $ref'd file, keyed so a file can be looked up by the File on any
// element's origin: recordingReader derives the key with the same rule
// kin-openapi uses for origin.Key.File (the full URL when absolute, otherwise the
// resolved filesystem path).
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
	// Origins may carry a "./" prefix on the same path (net/url prepends it
	// when a relative path's first segment contains a colon, as git refs do).
	key = strings.TrimPrefix(key, "./")
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
func recordingReader(inner openapi3.ReadFromURIFunc, capture *sourceCapture) openapi3.ReadFromURIFunc {
	return func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
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
