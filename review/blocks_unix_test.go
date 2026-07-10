//go:build unix

package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
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

// Composed sets commonly repeat a basename across directories (one
// openapi.yaml per service). Same-named components in same-named files stay
// separate blocks, each sliced from its own file; the block key is qualified
// with a directory segment so the two don't merge.
func TestExtract_ComposedSameBasenameStaysSeparate(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	base, err := load.NewSpecInfoFromGlobWithCapture(loader, "../data/review/samebase/base/*/openapi.yaml")
	require.NoError(t, err)
	rev, err := load.NewSpecInfoFromGlobWithCapture(loader, "../data/review/samebase/revision/*/openapi.yaml")
	require.NoError(t, err)

	var baseDocs, revDocs []*openapi3.T
	baseTexts, revTexts := map[string]string{}, map[string]string{}
	for _, si := range base {
		baseDocs = append(baseDocs, si.Spec)
		for k, v := range si.Sources {
			baseTexts[k] = v
		}
	}
	for _, si := range rev {
		revDocs = append(revDocs, si.Spec)
		for k, v := range si.Sources {
			revTexts[k] = v
		}
	}

	d, osm, err := diff.GetPathsDiff(diff.NewConfig(), base, rev)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(checker.NewConfig(checker.GetAllChecks()), d, osm, checker.INFO)

	blocks := Extract(changes, baseDocs, revDocs, baseTexts, revTexts)

	var a, b *Block
	for i := range blocks {
		switch {
		case strings.HasPrefix(blocks[i].Key, "svc-a/"):
			a = &blocks[i]
		case strings.HasPrefix(blocks[i].Key, "svc-b/"):
			b = &blocks[i]
		}
	}
	require.NotNil(t, a, "svc-a's User is its own block")
	require.NotNil(t, b, "svc-b's User is its own block")
	require.Equal(t, "components/schemas/User", a.Title)
	require.Equal(t, "components/schemas/User", b.Title)

	// each block slices its own file's schema, on both sides
	require.Contains(t, a.BaseText, "name:")
	require.NotContains(t, a.BaseText, "id:")
	require.Contains(t, a.RevText, "name:")
	require.Contains(t, b.BaseText, "id:")
	require.NotContains(t, b.BaseText, "name:")
	require.Contains(t, b.RevText, "extra:")
}
