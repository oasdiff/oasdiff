//go:build unix

package load

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewSpecInfoWithCapture records the root spec and every $ref'd file, keyed by
// resolved path; the keys match the File reported on each element's origin.
// Unix-only: it asserts on OS-native $ref path resolution, which uses a
// different separator on Windows.
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
