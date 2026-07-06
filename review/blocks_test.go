package review

import (
	"net/url"
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

// singleSpan returns the unique span for key, failing if absent or ambiguous.
func singleSpan(t *testing.T, idx docIndex, key string) span {
	t.Helper()
	entries := idx.byKey[key]
	require.Len(t, entries, 1, "expected exactly one span for %q", key)
	return entries[0]
}

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
	userSpan := singleSpan(t, buildIndex(doc), "components/schemas/User")

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
// survives the flatten.
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
		s := singleSpan(t, idx, name)
		require.Positive(t, s.start)
		require.GreaterOrEqual(t, s.end, s.start)
	}
	// Sections don't overlap and stay above the paths block.
	require.Less(t, singleSpan(t, idx, "info").end, singleSpan(t, idx, "servers").start)
	require.Less(t, singleSpan(t, idx, "security").start, singleSpan(t, idx, "GET /x").start)
}

// A security change (a SecurityChange, not an ApiChange) sourced in the security
// section cards to that section with its slice. This exercises both the section
// index and the interface-based source lookup that resolves non-ApiChange types.
func TestExtract_SecurityChangeCardsToSection(t *testing.T) {
	doc := loadWithOrigin(t, topLevelSpec)
	sec := singleSpan(t, buildIndex(doc), "security")

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
	oauth := singleSpan(t, idx, "components/securitySchemes/OAuth")

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

// loadWithOriginNamed loads an in-memory spec under a filename, so element
// origins carry that file (as composed/multi-file loads do).
func loadWithOriginNamed(t *testing.T, name, data string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromDataWithPath([]byte(data), &url.URL{Path: name})
	require.NoError(t, err)
	return doc
}

const dupUsersSpec = `openapi: 3.0.0
info: { title: users, version: "1" }
paths:
  /users:
    get:
      responses:
        "200": { description: ok }
components:
  schemas:
    Error:
      type: object
      properties:
        code: { type: integer }
`

const dupPetsBase = `openapi: 3.0.0
info: { title: pets, version: "1" }
paths:
  /pets:
    get:
      responses:
        "200": { description: ok }
components:
  schemas:
    Error:
      type: object
      properties:
        code: { type: string }
`

const dupPetsRev = `openapi: 3.0.0
info: { title: pets, version: "1" }
paths:
  /pets:
    get:
      responses:
        "200": { description: ok }
components:
  schemas:
    Error:
      type: object
      properties:
        code: { type: boolean }
`

// Composed specs can define the same component name in several files: the
// change slices the file it lives in and the card key is qualified by filename.
func TestExtract_ComposedDuplicateComponentName(t *testing.T) {
	usersBase := loadWithOriginNamed(t, "users.yaml", dupUsersSpec)
	usersRev := loadWithOriginNamed(t, "users.yaml", dupUsersSpec)
	petsBase := loadWithOriginNamed(t, "pets.yaml", dupPetsBase)
	petsRev := loadWithOriginNamed(t, "pets.yaml", dupPetsRev)

	// The change is inside pets.yaml's Error schema on both sides.
	baseSpan := buildIndex(petsBase).byKey["components/schemas/Error"][0]
	revSpan := buildIndex(petsRev).byKey["components/schemas/Error"][0]
	c := checker.ApiChange{
		Id:        "response-property-type-changed",
		Operation: "GET",
		Path:      "/pets",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{File: "pets.yaml", Line: baseSpan.end},
			RevisionSource: &checker.Source{File: "pets.yaml", Line: revSpan.end},
		},
	}

	blocks := Extract(checker.Changes{c},
		docs(usersBase, petsBase), docs(usersRev, petsRev),
		map[string]string{"users.yaml": dupUsersSpec, "pets.yaml": dupPetsBase},
		map[string]string{"users.yaml": dupUsersSpec, "pets.yaml": dupPetsRev})
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "pets.yaml: components/schemas/Error", b.Key, "the block key is qualified by file")
	require.Equal(t, "components/schemas/Error", b.Title, "the title stays plain; consumers label the file from BaseFile/RevFile")
	require.Equal(t, "pets.yaml", b.BaseFile)
	require.Equal(t, "pets.yaml", b.RevFile)
	require.Contains(t, b.BaseText, "string", "sliced from pets.yaml's Error, not users.yaml's")
	require.Contains(t, b.RevText, "boolean")
	require.NotContains(t, b.BaseText, "integer", "users.yaml's Error must not leak into the slice")
	require.NotContains(t, b.RevText, "integer")
}

// A git-revision span file keeps its "<rev>:" prefix while checker sources are
// display paths with the prefix stripped; matching must accept both forms.
func TestSmallestContaining_GitRevisionPrefixedFile(t *testing.T) {
	idx := docIndex{spans: []span{{key: "GET /users", title: "GET /users", file: "HEAD:api/openapi.yaml", start: 10, end: 20}}}

	s, ok := idx.smallestContaining("api/openapi.yaml", 12)
	require.True(t, ok, "a display-path source must match its rev-prefixed span")
	require.Equal(t, "GET /users", s.key)

	_, ok = idx.smallestContaining("other/openapi.yaml", 12)
	require.False(t, ok, "a different file must not match")
}

// Origins under git loads carry a "./" prefix that capture keys don't.
func TestTextFor_DotSlashOriginPrefix(t *testing.T) {
	texts := map[string]string{"HEAD:api.yaml": "x"}
	require.Equal(t, "x", textFor(texts, "./HEAD:api.yaml"))
	require.Equal(t, "x", textFor(texts, "HEAD:api.yaml"))
	require.Empty(t, textFor(texts, "other.yaml"))
}

const namedComponentsSpec = `openapi: 3.0.0
info: { title: t, version: "1.0" }
paths:
  /users:
    get:
      parameters:
        - $ref: '#/components/parameters/Limit'
      responses:
        "404": { $ref: '#/components/responses/NotFound' }
        "200":
          description: ok
          headers:
            X-Rate: { $ref: '#/components/headers/RateLimit' }
    post:
      requestBody: { $ref: '#/components/requestBodies/UserBody' }
      responses:
        "201": { description: created }
components:
  parameters:
    Limit:
      name: limit
      in: query
      schema: { type: integer }
  responses:
    NotFound:
      description: not found
      content:
        application/json:
          schema: { type: string }
  requestBodies:
    UserBody:
      content:
        application/json:
          schema: { type: object }
  headers:
    RateLimit:
      schema: { type: integer }
`

// Every named component type is indexed, so a change sourced inside one keys
// to its own block instead of falling back to the operation.
func TestBuildIndex_NamedComponentTypes(t *testing.T) {
	idx := buildIndex(loadWithOrigin(t, namedComponentsSpec))
	for _, key := range []string{
		"components/parameters/Limit",
		"components/responses/NotFound",
		"components/requestBodies/UserBody",
		"components/headers/RateLimit",
	} {
		s := singleSpan(t, idx, key)
		require.Positive(t, s.start, key)
		require.GreaterOrEqual(t, s.end, s.start, key)
	}
}

// The common case end to end: a change inside a $ref'd component response
// slices the response block, not the referencing operation.
func TestExtract_ComponentResponseBlock(t *testing.T) {
	doc := loadWithOrigin(t, namedComponentsSpec)
	notFound := singleSpan(t, buildIndex(doc), "components/responses/NotFound")

	c := checker.ApiChange{
		Id:        "response-property-type-changed",
		Operation: "GET",
		Path:      "/users",
		CommonChange: checker.CommonChange{
			BaseSource:     &checker.Source{Line: notFound.end},
			RevisionSource: &checker.Source{Line: notFound.end},
		},
	}
	blocks := Extract(checker.Changes{c}, docs(doc), docs(doc), oneFile("", namedComponentsSpec), oneFile("", namedComponentsSpec))
	require.Len(t, blocks, 1)
	require.Equal(t, "components/responses/NotFound", blocks[0].Key)
	require.Contains(t, blocks[0].BaseText, "NotFound:")
	require.NotContains(t, blocks[0].BaseText, "/users:", "not the operation block")
	require.NotContains(t, blocks[0].BaseText, "UserBody:", "sibling components excluded")
}
