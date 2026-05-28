//go:build unix

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

const minimalSpec = `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`

// TestLoadInfo_GitRevision creates a real git repo, commits a spec, and verifies
// that NewSpecInfo can load it via git revision syntax (e.g. "HEAD:openapi.yaml").
func TestLoadInfo_GitRevision(t *testing.T) {
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

	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(minimalSpec), 0644))

	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "add spec")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, "1.0", specInfo.GetVersion())
}

func TestLoadInfo_GitRevisionNotFound(t *testing.T) {
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

	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(minimalSpec), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "add spec")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	_, err = load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("HEAD:nonexistent.yaml"))
	require.ErrorContains(t, err, "failed to load spec from git revision")
}

// TestLoadInfo_TwoGitRevisionsSharedLoader verifies that loading two different git revisions
// of the same file with the same loader returns distinct specs.
// This guards against the openapi3 loader's visitedDocuments cache returning the cached
// first spec for the second load when both refs resolve to the same file path.
func TestLoadInfo_TwoGitRevisionsSharedLoader(t *testing.T) {
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

	specV1 := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`
	specV2 := `openapi: "3.0.0"
info:
  title: Test
  version: "2.0"
paths: {}
`
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(specV1), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "v1")

	require.NoError(t, os.WriteFile(specPath, []byte(specV2), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "v2")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	loader := openapi3.NewLoader()
	s1, err := load.NewSpecInfo(loader, load.NewSource("HEAD~1:openapi.yaml"))
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(loader, load.NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)

	require.Equal(t, "1.0", s1.GetVersion(), "base spec should be v1")
	require.Equal(t, "2.0", s2.GetVersion(), "revision spec should be v2")
}

// TestLoadInfo_GitRevisionThenLocalFile verifies that loading a git-revision spec followed
// by a local-file spec on the same loader works correctly (ReadFromURIFunc must not leak).
func TestLoadInfo_GitRevisionThenLocalFile(t *testing.T) {
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

	specV1 := minimalSpec
	specV2 := `openapi: "3.0.0"
info:
  title: Test
  version: "2.0"
paths: {}
`
	specPath := filepath.Join(dir, "openapi.yaml")
	localPath := filepath.Join(dir, "openapi-local.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(specV1), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "v1")

	require.NoError(t, os.WriteFile(localPath, []byte(specV2), 0644))

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := load.NewSpecInfo(loader, load.NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, "1.0", s1.GetVersion())

	s2, err := load.NewSpecInfo(loader, load.NewSource(localPath))
	require.NoError(t, err, "local-file load must not use the git ReadFromURIFunc")
	require.Equal(t, "2.0", s2.GetVersion())
	require.Nil(t, loader.ReadFromURIFunc, "loadFromGitRevision must not mutate the caller's loader")
}

// TestLoadInfo_GitRevisionSlashedRef_WithExternalRef verifies that $ref resolution
// works when the git revision contains a slash (e.g. "origin/main:openapi.yaml").
// Reproduces the production failure in oasdiff-test/tmp: path.Dir("origin/main:openapi.yaml")
// returns "origin" (not "origin/main:"), so a naive join mangles "./schemas/pet.yaml"
// into "origin/schemas/pet.yaml" instead of "origin/main:schemas/pet.yaml".
func TestLoadInfo_GitRevisionSlashedRef_WithExternalRef(t *testing.T) {
	dir := t.TempDir()
	remoteDir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	// Bare remote to enable "origin/main:..." refs.
	bareCmd := exec.Command("git", "init", "--bare", remoteDir)
	out, err := bareCmd.CombinedOutput()
	require.NoError(t, err, string(out))

	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")
	run("git", "remote", "add", "origin", remoteDir)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "schemas"), 0755))

	topLevel := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "./schemas/pet.yaml"
`
	petSchema := `type: object
required:
  - id
properties:
  id:
    type: integer
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(topLevel), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "schemas", "pet.yaml"), []byte(petSchema), 0644))

	run("git", "add", ".")
	run("git", "commit", "-m", "add multi-file spec")
	run("git", "push", "origin", "main")
	run("git", "fetch", "origin")

	// Remove working-tree copies so $ref resolution must go through "git show".
	require.NoError(t, os.Remove(filepath.Join(dir, "openapi.yaml")))
	require.NoError(t, os.Remove(filepath.Join(dir, "schemas", "pet.yaml")))

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("origin/main:openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, "1.0", specInfo.GetVersion())

	schema := specInfo.Spec.Paths.Value("/pets").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.NotNil(t, schema.Properties["id"], "id property from $ref chain should be resolved")
}

func TestLoadInfo_GitRevisionNoGit(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // remove git from PATH
	_, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("HEAD:openapi.yaml"))
	require.ErrorContains(t, err, "is git installed and in PATH?")
}

// TestLoadInfo_BlobHashRef verifies that the `<blob-hex>:<path>` form loads the
// blob's content directly, which is the form git's external-diff driver passes
// (git sends blob hashes, not commit hashes). The path portion is preserved on
// the ref for display purposes but not used to navigate a tree.
func TestLoadInfo_BlobHashRef(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return string(out)
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")

	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(minimalSpec), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "add spec")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	// Get the blob hash for the committed openapi.yaml.
	blobHex := run("git", "rev-parse", "HEAD:openapi.yaml")
	blobHex = blobHex[:len(blobHex)-1] // strip trailing newline
	require.Len(t, blobHex, 40, "git rev-parse should return a 40-char SHA-1")

	specInfo, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource(blobHex+":openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, "1.0", specInfo.GetVersion())
}

// TestLoadInfo_BlobHashRef_TwoBlobsSharedLoader verifies that loading two
// different blob hashes for the same path on the same loader returns distinct
// specs. This is the exact shape git invokes the external-diff driver with:
// old and new blob hashes, both notionally at the same in-tree path. Without
// unique cache keys per ref the loader would return the cached first spec for
// the second load.
func TestLoadInfo_BlobHashRef_TwoBlobsSharedLoader(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return string(out)
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")

	specV1 := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths: {}
`
	specV2 := `openapi: "3.0.0"
info:
  title: Test
  version: "2.0"
paths: {}
`
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(specV1), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "v1")

	require.NoError(t, os.WriteFile(specPath, []byte(specV2), 0644))
	run("git", "add", "openapi.yaml")
	run("git", "commit", "-m", "v2")

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	baseHex := run("git", "rev-parse", "HEAD~1:openapi.yaml")
	baseHex = baseHex[:len(baseHex)-1]
	revHex := run("git", "rev-parse", "HEAD:openapi.yaml")
	revHex = revHex[:len(revHex)-1]

	loader := openapi3.NewLoader()
	s1, err := load.NewSpecInfo(loader, load.NewSource(baseHex+":openapi.yaml"))
	require.NoError(t, err)
	s2, err := load.NewSpecInfo(loader, load.NewSource(revHex+":openapi.yaml"))
	require.NoError(t, err)

	require.Equal(t, "1.0", s1.GetVersion(), "base blob should be v1")
	require.Equal(t, "2.0", s2.GetVersion(), "revision blob should be v2")
}

// TestLoadInfo_GitRevisionWithExternalRef verifies that $ref chains are resolved via
// "git show" when loading from a git revision, not opened as literal file paths.
func TestLoadInfo_GitRevisionWithExternalRef(t *testing.T) {
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

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "schemas"), 0755))

	topLevel := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "./schemas/pet.yaml"
`
	petSchema := `type: object
required:
  - id
properties:
  id:
    type: integer
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(topLevel), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "schemas", "pet.yaml"), []byte(petSchema), 0644))

	run("git", "add", ".")
	run("git", "commit", "-m", "add multi-file spec")

	// Remove the files after committing so they only exist in git history.
	// This ensures $ref resolution must go through "git show" — if the
	// repo-root gitPrefix restoration fails, the relative ref resolves to
	// "schemas/pet.yaml" (no "HEAD:" prefix) and the load will fail.
	require.NoError(t, os.Remove(filepath.Join(dir, "openapi.yaml")))
	require.NoError(t, os.Remove(filepath.Join(dir, "schemas", "pet.yaml")))

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldDir) //nolint:errcheck

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	specInfo, err := load.NewSpecInfo(loader, load.NewSource("HEAD:openapi.yaml"))
	require.NoError(t, err)
	require.Equal(t, "1.0", specInfo.GetVersion())

	// Verify the $ref was resolved — the pet schema should have the "id" property
	schema := specInfo.Spec.Paths.Value("/pets").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	require.NotNil(t, schema.Properties["id"], "id property from $ref chain should be resolved")
}
