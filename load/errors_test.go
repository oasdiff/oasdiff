package load_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// kin-openapi refuses an external $ref itself when no custom ReadFromURIFunc is
// installed, and its error carries no type to match on. The load runs against
// the real loader so a change to kin's wording fails here rather than silently
// turning a refusal back into a plain load failure.
func TestIsExternalRefDisallowed_KinRefusal(t *testing.T) {
	for name, ref := range map[string]string{
		"whole file":        "target.yaml",
		"file and fragment": "target.yaml#/Leak",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			spec := filepath.Join(dir, "openapi.yaml")
			require.NoError(t, os.WriteFile(spec, []byte(`openapi: 3.0.0
info: {title: t, version: "1.0"}
paths:
  /a:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: {$ref: "`+ref+`"}
`), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "target.yaml"), []byte("Leak: {type: string}\n"), 0o644))

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = false
			_, err := load.NewSpecInfo(loader, load.NewSource(spec))

			require.Error(t, err)
			require.True(t, load.IsExternalRefDisallowed(err), "kin refused the $ref: %v", err)
		})
	}
}

// A load that fails for any other reason must not be read as a refusal, or
// every unreadable spec would report the external-refs remedy.
func TestIsExternalRefDisallowed_OtherFailures(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	_, err := load.NewSpecInfo(loader, load.NewSource(filepath.Join(t.TempDir(), "missing.yaml")))

	require.Error(t, err)
	require.False(t, load.IsExternalRefDisallowed(err))
	require.False(t, load.IsExternalRefDisallowed(nil))
}
