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

// An external $ref authored "other.yaml" on the base side and "./other.yaml" on
// the revision side names the same target; kin preserves the authored "./" in
// sr.Ref, so indexExternalSchemas strips it before keying. That makes both sides
// key the block identically, so resolve() pairs them into one block rather than a
// spurious delete + add. Remove the strip and this fails: base and rev key
// differently and the base side no longer slices.
func TestExtract_CrossFileRefStyleDiffersBetweenSides(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.yaml"), []byte(crossFileSubDoc), 0644))
	baseRoot := filepath.Join(dir, "base.yaml")
	revRoot := filepath.Join(dir, "rev.yaml")
	require.NoError(t, os.WriteFile(baseRoot, []byte(strings.Replace(crossFileRoot, "./other.yaml", "other.yaml", 1)), 0644))
	require.NoError(t, os.WriteFile(revRoot, []byte(crossFileRoot), 0644)) // keeps "./other.yaml"

	loadSpec := func(path string) *load.SpecInfo {
		loader := openapi3.NewLoader()
		loader.IncludeOrigin = true
		loader.IsExternalRefsAllowed = true
		si, err := load.NewSpecInfoWithCapture(loader, load.NewSource(path))
		require.NoError(t, err)
		return si
	}
	baseSI, revSI := loadSpec(baseRoot), loadSpec(revRoot)

	findExternal := func(d *openapi3.T) (string, span) {
		for k, spans := range buildIndex(d).byKey {
			if strings.Contains(k, "other.yaml") {
				return k, spans[0]
			}
		}
		return "", span{}
	}
	baseKey, baseSpan := findExternal(baseSI.Spec)
	revKey, revSpan := findExternal(revSI.Spec)
	require.NotEmpty(t, baseKey, "the external User schema must be indexed on the base side")
	require.Equal(t, baseKey, revKey, "'other.yaml' and './other.yaml' must key the same block")

	// one change, sourced inside other.yaml on each side (as the changelog reports)
	c := checker.ApiChange{
		Id:        "request-property-type-changed",
		Operation: "POST",
		Path:      "/users",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{File: baseSpan.file, Line: baseSpan.end},
			RevisionSource: &checker.Source{File: revSpan.file, Line: revSpan.end},
		},
	}
	blocks := Extract(checker.Changes{c}, docs(baseSI.Spec), docs(revSI.Spec), baseSI.Sources, revSI.Sources)
	require.Len(t, blocks, 1, "the two sides pair into one block")
	require.NotEmpty(t, blocks[0].BaseText, "base side sliced (paired, not delete + add)")
	require.NotEmpty(t, blocks[0].RevText, "revision side sliced")
}
