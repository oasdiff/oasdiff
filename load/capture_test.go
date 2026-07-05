package load

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

const captureSubDoc = `openapi: 3.0.0
info: { title: lib, version: "1" }
paths: {}
components:
  schemas:
    User: { type: object }
`

const captureRoot = `openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: { $ref: './other.yaml#/components/schemas/User' }
`

func captureLoader() *openapi3.Loader {
	l := openapi3.NewLoader()
	l.IncludeOrigin = true
	l.IsExternalRefsAllowed = true
	return l
}

// NewSpecInfoWithCapture records the root spec and every $ref'd file, keyed by
// resolved path; the keys match the File reported on each element's origin.
func TestNewSpecInfoWithCapture_RecordsAllFiles(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "openapi.yaml")
	other := filepath.Join(dir, "other.yaml")
	require.NoError(t, os.WriteFile(root, []byte(captureRoot), 0644))
	require.NoError(t, os.WriteFile(other, []byte(captureSubDoc), 0644))

	si, err := NewSpecInfoWithCapture(captureLoader(), NewSource(root))
	require.NoError(t, err)
	require.Contains(t, si.Sources, root, "the root file is captured")
	require.Contains(t, si.Sources, other, "the $ref'd file is captured")
	require.Equal(t, captureSubDoc, si.Sources[other], "captured content is the file's text verbatim")

	// The keys are the origin Files, so a consumer can slice by origin file.
	user := si.Spec.Paths.Value("/x").Get.Responses.Value("200").Value.Content["application/json"].Schema
	require.Equal(t, other, user.Value.Origin.Key.File, "capture key == resolved element's origin File")
}

// A plain load installs no recorder and leaves Sources nil.
func TestNewSpecInfo_NoCapture(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "openapi.yaml")
	other := filepath.Join(dir, "other.yaml")
	require.NoError(t, os.WriteFile(root, []byte(captureRoot), 0644))
	require.NoError(t, os.WriteFile(other, []byte(captureSubDoc), 0644))

	si, err := NewSpecInfo(captureLoader(), NewSource(root))
	require.NoError(t, err)
	require.Nil(t, si.Sources, "plain load does not capture sources")
}

// A URL-loaded spec must be captured under its full URL, matching the origin
// File kin records for remote documents.
func TestNewSpecInfoWithCapture_URL(t *testing.T) {
	const spec = "openapi: 3.0.0\ninfo: {title: t, version: \"1\"}\npaths: {}\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(spec))
	}))
	defer srv.Close()

	u := srv.URL + "/openapi.yaml"
	si, err := NewSpecInfoWithCapture(openapi3.NewLoader(), NewSource(u))
	require.NoError(t, err)
	require.Equal(t, spec, si.Sources[u], "captured under the full URL")
}
