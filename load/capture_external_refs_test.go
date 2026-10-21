package load_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// The capture reader that --open installs is a custom ReadFromURIFunc, and
// kin-openapi applies IsExternalRefsAllowed only when no custom reader is
// installed. A reader that does not apply it turns --allow-external-refs=false
// into a no-op, letting a spec read any local file or reach any URL and putting
// what it read into the uploaded review.
func TestCapture_HonorsExternalRefsPolicy(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	spec := filepath.Join(dir, "spec")
	require.NoError(t, os.MkdirAll(outside, 0o755))
	require.NoError(t, os.MkdirAll(spec, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "target.yaml"),
		[]byte("Leak:\n  type: string\n  format: OUTSIDE\n"), 0o644))

	write := func(name, ref string) string {
		p := filepath.Join(spec, name)
		require.NoError(t, os.WriteFile(p, []byte(`openapi: 3.0.0
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
                $ref: "`+ref+`"
`), 0o644))
		return p
	}

	// Each case names its own spec file: the cases share a directory, so
	// deriving the name from the case would let two of them collide and
	// overwrite each other's spec.
	for _, tt := range []struct {
		name string
		file string
		ref  string
	}{
		{"outside the spec directory", "traversal.yaml", "../outside/target.yaml#/Leak"},
		{"sibling file", "sibling_ref.yaml", "sibling.yaml#/Leak"},
		{"http url", "http_ref.yaml", "http://127.0.0.1:1/target.yaml#/Leak"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := write(tt.file, tt.ref)

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = false
			_, err := load.NewSpecInfoWithCapture(loader, load.NewSource(path))

			require.Error(t, err, "the capture reader must apply the policy")
			var extErr *load.ExternalRefError
			require.ErrorAs(t, err, &extErr, "callers map this type to a dedicated exit code")
		})
	}
}

// With external refs allowed, --open keeps working and still captures the
// referenced file, which is what the review bundle slices blocks from.
func TestCapture_AllowedExternalRefStillCaptured(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "target.yaml"),
		[]byte("Leak:\n  type: string\n  format: OUTSIDE\n"), 0o644))
	path := filepath.Join(dir, "rev.yaml")
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
                $ref: "target.yaml#/Leak"
`), 0o644))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	si, err := load.NewSpecInfoWithCapture(loader, load.NewSource(path))
	require.NoError(t, err)
	require.Len(t, si.Sources, 2, "root and the referenced file")
}

// The git-revision reader is the one custom reader that passes no root to
// withCapture, so allowLocation returns early for it and its own
// IsExternalRefsAllowed check is the only thing refusing an external $ref.
// With --open on top of a git source, both readers are in play and that single
// check has to survive the composition. Nothing else covers it: the non-capture
// git path has TestLoadInfo_GitRevisionExternalRefBlocked, and the capture
// tests above all load a plain file.
func TestCapture_GitRevisionHonorsExternalRefsPolicy(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(`openapi: 3.0.0
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
                $ref: "http://127.0.0.1:1/target.yaml#/Leak"
`), 0o644))
	run("git", "add", ".")
	run("git", "commit", "-m", "spec with an external ref")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	_, err = load.NewSpecInfoWithCapture(loader, load.NewSource("HEAD:openapi.yaml"))

	require.Error(t, err, "capture must not weaken the git path's policy check")
	var extErr *load.ExternalRefError
	require.ErrorAs(t, err, &extErr, "callers map this type to a dedicated exit code")
}

// Composed mode loads each glob member with its own capture and its own root,
// so the policy has to hold per member rather than only for the first.
func TestCapture_GlobHonorsExternalRefsPolicy(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "target.yaml"),
		[]byte("Leak:\n  type: string\n  format: OUTSIDE\n"), 0o644))

	write := func(name, path, schema string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(`openapi: 3.0.0
info: {title: t, version: "1.0"}
paths:
  `+path+`:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                `+schema+`
`), 0o644))
	}

	// The clean member comes first in glob order, so a policy applied only to
	// the first member would let the second one through.
	write("api-1.yaml", "/a", `type: string`)
	write("api-2.yaml", "/b", `$ref: "target.yaml#/Leak"`)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	_, err := load.NewSpecInfoFromGlobWithCapture(loader, filepath.Join(dir, "api-*.yaml"))

	require.Error(t, err, "every glob member must apply the policy")
	var extErr *load.ExternalRefError
	require.ErrorAs(t, err, &extErr, "callers map this type to a dedicated exit code")
}
