package load

import (
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
