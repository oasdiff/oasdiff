package load

import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// A refPolicy that names neither a root document nor an enforcing inner reader
// matches no document, so it refuses every $ref. That is what a call site gets
// by not choosing, so the omission fails closed.
func TestRefPolicy_ZeroValueFailsClosed(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	location, err := url.Parse("http://169.254.169.254/latest/meta-data/schema.yaml")
	require.NoError(t, err)

	err = refPolicy{}.checkLocation(loader, location)

	require.Error(t, err, "a policy that names no root must refuse every $ref")
	var extErr *ExternalRefError
	require.ErrorAs(t, err, &extErr)
}

// The inner reader owning the policy is the one case that legitimately skips
// the comparison, and it has to be chosen by name.
func TestRefPolicy_InnerEnforcesSkipsTheComparison(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	location, err := url.Parse("HEAD:openapi.yaml")
	require.NoError(t, err)

	require.NoError(t, refPolicy{innerEnforces: true}.checkLocation(loader, location),
		"the git reader applies the policy itself")
}

// A URL is the root document only when it is the root as given. Its path
// equalling the root's path is not enough: any host can serve that path.
func TestSameLocation(t *testing.T) {
	absRoot, err := filepath.Abs("openapi.yaml")
	require.NoError(t, err)
	// a URL path starts with a slash, which a Windows path does not
	urlPath := func(path string) string {
		return "/" + strings.TrimPrefix(filepath.ToSlash(path), "/")
	}
	// kin-openapi hands the reader a file as a URL that has a path only
	file := func(path string) *url.URL {
		return &url.URL{Path: filepath.ToSlash(path)}
	}
	parse := func(rawURL string) *url.URL {
		location, err := url.Parse(rawURL)
		require.NoError(t, err)
		return location
	}

	const urlRoot = "https://example.com/openapi.yaml"
	// the path filepath.Abs gives a URL root, which a host can serve too
	absURLRoot, err := filepath.Abs(urlRoot)
	require.NoError(t, err)

	for _, tt := range []struct {
		name     string
		location *url.URL
		root     string
		same     bool
	}{
		{"file root as given", file("openapi.yaml"), "openapi.yaml", true},
		{"file root in its absolute form", file(absRoot), "openapi.yaml", true},
		{"file root uncleaned", file("./openapi.yaml"), "openapi.yaml", true},
		{"another file", file("other.yaml"), "openapi.yaml", false},
		{"url root as given", parse(urlRoot), urlRoot, true},
		{"url with the path of an absolute file root", parse("http://other.example.com" + urlPath(absRoot)), absRoot, false},
		{"url with the path of a relative file root", parse("http://other.example.com" + urlPath(absRoot)), "openapi.yaml", false},
		{"host without a scheme with the path of the file root", parse("//other.example.com" + urlPath(absRoot)), absRoot, false},
		{"url with the path of a url root", parse("http://other.example.com/openapi.yaml"), urlRoot, false},
		{"url with the resolved path of a url root", parse("http://other.example.com" + urlPath(absURLRoot)), urlRoot, false},
		{"url root with the scheme changed", parse("http://example.com/openapi.yaml"), urlRoot, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.same, sameLocation(tt.location, tt.root))
		})
	}
}
