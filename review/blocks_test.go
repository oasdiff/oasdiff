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

// oneFile builds the file->text map Extract takes for a single-file (or
// in-memory, file "") spec.
func oneFile(file, text string) map[string]string { return map[string]string{file: text} }

// docs wraps specs for Extract (which takes a set of docs per side for composed).
func docs(d ...*openapi3.T) []*openapi3.T { return d }

const endpointSpec = `openapi: 3.0.0
info:
  title: t
  version: "1"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200": { description: ok }
    post:
      operationId: createUser
      responses:
        "201": { description: created }
  /health:
    get:
      operationId: health
      responses:
        "200": { description: ok }
`

func loadWithOrigin(t *testing.T, data string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(data))
	require.NoError(t, err)
	return doc
}

// Extract slices the exact operation block for an endpoint change: just the
// POST operation, not the sibling GET, not the other path, not the whole doc.
func TestExtract_OperationBlock(t *testing.T) {
	doc := loadWithOrigin(t, endpointSpec)
	changes := checker.Changes{checker.ApiChange{Id: "c1", Operation: "POST", Path: "/users"}}

	blocks := Extract(changes, docs(doc), docs(doc), oneFile("", endpointSpec), oneFile("", endpointSpec))
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "POST /users", b.Key)
	require.Equal(t, []string{"c1"}, b.ChangeIDs)
	require.Positive(t, b.BaseLineStart)
	require.Contains(t, b.BaseText, "post:")
	require.Contains(t, b.BaseText, "operationId: createUser")
	require.NotContains(t, b.BaseText, "operationId: listUsers") // sibling GET excluded
	require.NotContains(t, b.BaseText, "/health:")               // other path excluded
	require.NotContains(t, b.BaseText, "openapi:")               // not the whole document
	require.Equal(t, b.BaseText, b.RevText)                      // identical specs here
}

// A path-level change (no method) slices the whole path block, all methods.
func TestExtract_PathBlock(t *testing.T) {
	doc := loadWithOrigin(t, endpointSpec)
	changes := checker.Changes{checker.ApiChange{Id: "c1", Path: "/users"}}

	blocks := Extract(changes, docs(doc), docs(doc), oneFile("", endpointSpec), oneFile("", endpointSpec))
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "/users", b.Key)
	require.Contains(t, b.BaseText, "/users:")
	require.Contains(t, b.BaseText, "operationId: listUsers") // both methods present
	require.Contains(t, b.BaseText, "operationId: createUser")
	require.NotContains(t, b.BaseText, "/health:") // other path excluded
}

func TestSliceLines(t *testing.T) {
	text := "l1\nl2\nl3\nl4\nl5"
	for _, tc := range []struct {
		name       string
		start, end int
		want       string
	}{
		{"single line", 2, 2, "l2"},
		{"range", 2, 4, "l2\nl3\nl4"},
		{"whole", 1, 5, "l1\nl2\nl3\nl4\nl5"},
		{"clamped end", 4, 99, "l4\nl5"},
		{"out of range start", 99, 100, ""},
		{"inverted", 4, 2, ""},
		{"zero start", 0, 2, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, sliceLines(text, tc.start, tc.end))
		})
	}
}

// fallbackKey is used when a change has no resolvable source line: key by
// operation+path, else path, else (via Area) a top-level bucket, else "Other".
func TestFallbackKey(t *testing.T) {
	for _, tc := range []struct {
		name           string
		op, path       string
		wantKey, wantT string
	}{
		{"operation", "POST", "/users", "POST /users", "POST /users"},
		{"path level", "", "/users", "/users", "/users"},
		{"no path no area", "", "", otherChangesKey, "Other changes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, title := fallbackKey(checker.ApiChange{Operation: tc.op, Path: tc.path})
			require.Equal(t, tc.wantKey, key)
			require.Equal(t, tc.wantT, title)
		})
	}
}

const refdComponentSpec = `openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema: { $ref: '#/components/schemas/User' }
      responses: { "201": { description: created } }
components:
  schemas:
    User:
      type: object
      properties:
        role: { type: string }
`

// The hard case: a change to a property inside a $ref'd component. oasdiff
// reports it under the referencing operation (POST /users), but its source line
// is inside components/schemas/User. The block must follow the source line and
// card it as the component (whose block holds the real change), not the
// operation (whose source only shows the $ref).
func TestExtract_RefdComponentFollowsSourceLine(t *testing.T) {
	doc := loadWithOrigin(t, refdComponentSpec)
	userSpan, ok := buildIndex(doc).byKey["components/schemas/User"]
	require.True(t, ok)

	// change reported under the endpoint, but sourced inside User (the role line)
	c := checker.ApiChange{
		Id:        "request-property-type-changed",
		Operation: "POST",
		Path:      "/users",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{Line: userSpan.end},
			RevisionSource: &checker.Source{Line: userSpan.end},
		},
	}

	blocks := Extract(checker.Changes{c}, docs(doc), docs(doc), oneFile("", refdComponentSpec), oneFile("", refdComponentSpec))
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "components/schemas/User", b.Key) // the component, not "POST /users"
	require.Contains(t, b.BaseText, "User:")
	require.Contains(t, b.BaseText, "role:")
	require.NotContains(t, b.BaseText, "$ref")    // not the operation block
	require.NotContains(t, b.BaseText, "/users:") // not the path block
}

// allOfRequiredBase / allOfRequiredRevision differ by one line: the revision
// requires `role` at the usage site via allOf, not inside the User component.
const allOfRequiredBase = `openapi: 3.0.0
info: { title: t, version: "1.0" }
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema:
              allOf:
                - $ref: '#/components/schemas/User'
      responses:
        "200": { description: ok }
components:
  schemas:
    User:
      type: object
      properties:
        name: { type: string }
        role: { type: string }
`

const allOfRequiredRevision = `openapi: 3.0.0
info: { title: t, version: "1.0" }
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema:
              allOf:
                - $ref: '#/components/schemas/User'
                - required: [role]
      responses:
        "200": { description: ok }
components:
  schemas:
    User:
      type: object
      properties:
        name: { type: string }
        role: { type: string }
`

// End-to-end through the real diff + checker pipeline with --flatten-allof.
// Flattening gives the clear message (request-property-became-required) but
// leaves the change with no source line, since the merged schema is a new
// construct. The change must still card onto the operation it names (the
// fallback), and the slice must be the real operation block whose origin
// survives the flatten. This is the path the review page runs.
func TestExtract_FlattenedAllOfFallsBackToOperation(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true

	s1, err := load.NewSpecInfoFromData(loader, []byte(allOfRequiredBase), "base.yaml", load.WithFlattenAllOf())
	require.NoError(t, err)
	s2, err := load.NewSpecInfoFromData(loader, []byte(allOfRequiredRevision), "revision.yaml", load.WithFlattenAllOf())
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibility(checker.NewConfig(checker.GetAllChecks()), d, sources)

	// flatten produces the readable, semantic change, with no source line
	require.Len(t, changes, 1)
	require.Equal(t, "request-property-became-required", changes[0].GetId())
	base, rev := changeSources(changes[0])
	require.Nil(t, base)
	require.Nil(t, rev)

	blocks := Extract(changes, docs(s1.Spec), docs(s2.Spec), oneFile("base.yaml", allOfRequiredBase), oneFile("revision.yaml", allOfRequiredRevision))
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "POST /users", b.Key) // fell back to the operation it names
	require.Contains(t, b.RevText, "required: [role]")
	require.Contains(t, b.RevText, "$ref: '#/components/schemas/User'")
	require.NotContains(t, b.RevText, "openapi:") // a real slice, not the whole doc
}

const topLevelSpec = `openapi: 3.0.0
info:
  title: t
  version: "1.0"
servers:
  - url: https://a.example.com
tags:
  - name: t1
security:
  - apiKey: []
paths:
  /x:
    get:
      responses:
        "200": { description: ok }
components:
  securitySchemes:
    apiKey: { type: apiKey, name: k, in: header }
`

// The four top-level sections are indexed, each spanning from its key line to
// just before the next top-level key.
func TestBuildIndex_TopLevelSections(t *testing.T) {
	idx := buildIndex(loadWithOrigin(t, topLevelSpec))
	for _, name := range []string{"info", "servers", "tags", "security"} {
		s, ok := idx.byKey[name]
		require.True(t, ok, "section %q must be indexed", name)
		require.Positive(t, s.start)
		require.GreaterOrEqual(t, s.end, s.start)
	}
	// Sections don't overlap and stay above the paths block.
	require.Less(t, idx.byKey["info"].end, idx.byKey["servers"].start)
	require.Less(t, idx.byKey["security"].start, idx.byKey["GET /x"].start)
}

// A security change (a SecurityChange, not an ApiChange) sourced in the security
// section cards to that section with its slice. This exercises both the section
// index and the interface-based source lookup that resolves non-ApiChange types.
func TestExtract_SecurityChangeCardsToSection(t *testing.T) {
	doc := loadWithOrigin(t, topLevelSpec)
	sec, ok := buildIndex(doc).byKey["security"]
	require.True(t, ok)

	c := checker.SecurityChange{
		Id:           "api-security-removed",
		CommonChange: checker.CommonChange{RevisionSource: &checker.Source{Line: sec.start}},
	}

	blocks := Extract(checker.Changes{c}, docs(doc), docs(doc), oneFile("", topLevelSpec), oneFile("", topLevelSpec))
	require.Len(t, blocks, 1)
	b := blocks[0]
	require.Equal(t, "security", b.Key, "a security change cards to the security section, not Other")
	require.Contains(t, b.RevText, "security:")
	require.Contains(t, b.RevText, "apiKey")
	require.NotContains(t, b.RevText, "/x:") // not the paths block
}

const crossFileSubDoc = `openapi: 3.0.0
info: { title: lib, version: "1" }
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        role: { type: string }
`

const crossFileRoot = `openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema: { $ref: './other.yaml#/components/schemas/User' }
      responses: { "201": { description: created } }
`

// A change inside a schema $ref'd from another file cards to a block keyed by
// the ref and slices from that external file, not the referencing operation in
// the root file. The texts come from the captured Sources, keyed by origin file.
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
	for k, s := range idx.byKey {
		if strings.Contains(k, "other.yaml") {
			userKey, userSpan = k, s
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

const securitySchemeSpec = `openapi: 3.0.0
info: { title: t, version: "1.0" }
paths: {}
components:
  securitySchemes:
    AccessToken:
      type: http
      scheme: bearer
    OAuth:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/auth
          tokenUrl: https://example.com/token
          scopes:
            read: read access
`

// A component security-scheme change (scheme removed / type changed) is sourced
// inside components/securitySchemes, which must be indexed so the change slices
// that scheme's block instead of falling back.
func TestExtract_SecuritySchemeBlock(t *testing.T) {
	doc := loadWithOrigin(t, securitySchemeSpec)
	idx := buildIndex(doc)
	oauth, ok := idx.byKey["components/securitySchemes/OAuth"]
	require.True(t, ok, "security schemes must be indexed")

	// change reported at the OAuth scheme's source line (as the checker sources it)
	c := checker.ApiChange{
		Id: "api-security-component-removed",
		CommonChange: checker.CommonChange{
			BaseSource: &checker.Source{Line: oauth.start},
		},
	}
	blocks := Extract(checker.Changes{c}, docs(doc), docs(doc), oneFile("", securitySchemeSpec), oneFile("", securitySchemeSpec))
	require.Len(t, blocks, 1)
	require.Equal(t, "components/securitySchemes/OAuth", blocks[0].Key)
	require.Contains(t, blocks[0].BaseText, "OAuth:")
	require.Contains(t, blocks[0].BaseText, "oauth2")
	require.NotContains(t, blocks[0].BaseText, "AccessToken:") // sibling scheme excluded
}
