package load_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// With external refs disallowed, a spec may still refer to itself, however
// the root document and the reference each spell its path. The two can differ
// by a "./" or a "..", which needs the cleaned comparison in sameLocation, and
// by one of them being absolute, which needs the resolved one.
func TestCapture_SelfReferenceIsNotExternal(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
	path := filepath.Join(dir, "openapi.yaml")
	t.Chdir(dir)

	for _, tt := range []struct {
		name   string
		source string
		ref    string
		// noVolume marks a reference that kin-openapi loads only when the
		// path has no drive letter: it parses "C:/..." as a URL with the
		// scheme "c", and reads the path of "file:///C:/..." as "/C:/...".
		noVolume bool
	}{
		{"fragment only", "openapi.yaml", "#/components/schemas/Leak", false},
		{"by file name", "openapi.yaml", "openapi.yaml#/components/schemas/Leak", false},
		{"by file name, root with a dot", "./openapi.yaml", "openapi.yaml#/components/schemas/Leak", false},
		{"by dotted name, plain root", "openapi.yaml", "./openapi.yaml#/components/schemas/Leak", false},
		{"by file name, root through a parent", "sub/../openapi.yaml", "openapi.yaml#/components/schemas/Leak", false},
		{"by file name, absolute root", path, "openapi.yaml#/components/schemas/Leak", false},
		{"by absolute path, relative root", "openapi.yaml", filepath.ToSlash(path) + "#/components/schemas/Leak", true},
		{"by absolute path, absolute root", path, filepath.ToSlash(path) + "#/components/schemas/Leak", false},
		{"by file url, relative root", "openapi.yaml", "file://" + filepath.ToSlash(path) + "#/components/schemas/Leak", true},
		{"by file url, absolute root", path, "file://" + filepath.ToSlash(path) + "#/components/schemas/Leak", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.noVolume && filepath.VolumeName(path) != "" {
				t.Skip("kin-openapi does not load this reference to a path with a drive letter")
			}
			require.NoError(t, os.WriteFile(path, []byte(`openapi: 3.0.0
info: {title: t, version: "1.0"}
paths:
  /a:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "`+tt.ref+`"
components:
  schemas:
    Leak:
      type: string
`), 0o644))

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = false
			_, err := load.NewSpecInfoWithCapture(loader, load.NewSource(tt.source))
			require.NoError(t, err)
		})
	}
}
