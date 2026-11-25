package load_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// With external refs disallowed, a spec may still refer to itself, however
// the root document and the reference each spell its path. The two can differ
// by a "./" or a "..", which needs the cleaned comparison in sameLocation, and
// by one of them being absolute, which needs the resolved one.
const drivePathFromRelativeRoot = "by absolute path, relative root"

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
	}{
		{"fragment only", "openapi.yaml", "#/components/schemas/Leak"},
		{"by file name", "openapi.yaml", "openapi.yaml#/components/schemas/Leak"},
		{"by file name, root with a dot", "./openapi.yaml", "openapi.yaml#/components/schemas/Leak"},
		{"by dotted name, plain root", "openapi.yaml", "./openapi.yaml#/components/schemas/Leak"},
		{"by file name, root through a parent", "sub/../openapi.yaml", "openapi.yaml#/components/schemas/Leak"},
		{"by file name, absolute root", path, "openapi.yaml#/components/schemas/Leak"},
		{drivePathFromRelativeRoot, "openapi.yaml", filepath.ToSlash(path) + "#/components/schemas/Leak"},
		{"by absolute path, absolute root", path, filepath.ToSlash(path) + "#/components/schemas/Leak"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && tt.name == drivePathFromRelativeRoot {
				t.Skip("kin-openapi parses a $ref to a drive path as a URL with the drive letter as its scheme, and does not load it")
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
