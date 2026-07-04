package load

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// sourceCapture records the raw bytes of every file the loader reads, the root
// spec and each $ref'd file, keyed by the path the loader resolved (which
// matches the File reported on each element's origin). It is populated only on
// the review-bundle path (NewSpecInfoWithCapture); ordinary loads install no
// recorder and pay nothing.
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
// to record every successfully read file into the capture. The key is the
// location's filesystem path, matching origin.Key.File.
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
			capture.record(filepath.FromSlash(location.Path), data)
		}
		return data, err
	}
}
