//go:build unix

package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// A change inside a schema $ref'd from another file cards to a block keyed by
// the ref and slices from that external file, not the referencing operation in
// the root file. The texts come from the captured Sources, keyed by origin file.
// Unix-only: it asserts on OS-native $ref path resolution, which uses a
// different separator on Windows.
func TestExtract_CrossFileSchemaSlicedFromExternalFile(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "openapi.yaml")
	other := filepath.Join(dir, "other.yaml")
	require.NoError(t, os.WriteFile(root, []byte(crossFileRoot), 0644))
	require.NoError(t, os.WriteFile(other, []byte(crossFileSubDoc), 0644))

	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	loader.IsExternalRefsAllowed = true
	si, err := load.NewSpecInfoWithCapture(loader, load.NewSource(root))
	require.NoError(t, err)

	// the cross-file User block is indexed, in other.yaml
	idx := buildIndex(si.Spec)
	var userKey string
	var userSpan span
	for k, spans := range idx.byKey {
		if strings.Contains(k, "other.yaml") {
			userKey, userSpan = k, spans[0]
		}
	}
	require.NotEmpty(t, userKey, "the cross-file User schema must be indexed")
	require.Equal(t, other, userSpan.file, "indexed in the external file")

	// a change sourced inside the external file (as the changelog reports it)
	c := checker.ApiChange{
		Id:           "request-property-type-changed",
		Operation:    "POST",
		Path:         "/users",
		CommonChange: checker.CommonChange{RevisionSource: &checker.Source{File: other, Line: userSpan.end}},
	}
	blocks := Extract(checker.Changes{c}, docs(si.Spec), docs(si.Spec), si.Sources, si.Sources)
	require.Len(t, blocks, 1)
	require.Equal(t, userKey, blocks[0].Key, "cards to the external block, not the operation")
	require.Equal(t, "other.yaml", blocks[0].RevFile, "the block reports the file it was sliced from")
	require.Contains(t, blocks[0].RevText, "User:")
	require.Contains(t, blocks[0].RevText, "role:")
	require.NotContains(t, blocks[0].RevText, "/users:", "sliced from other.yaml, not the root spec")
}
