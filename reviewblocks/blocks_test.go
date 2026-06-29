package reviewblocks

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

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

	blocks := Extract(changes, doc, doc, endpointSpec, endpointSpec)
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

	blocks := Extract(changes, doc, doc, endpointSpec, endpointSpec)
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

	blocks := Extract(checker.Changes{c}, doc, doc, refdComponentSpec, refdComponentSpec)
	require.Len(t, blocks, 1)
	b := blocks[0]

	require.Equal(t, "components/schemas/User", b.Key) // the component, not "POST /users"
	require.Contains(t, b.BaseText, "User:")
	require.Contains(t, b.BaseText, "role:")
	require.NotContains(t, b.BaseText, "$ref")    // not the operation block
	require.NotContains(t, b.BaseText, "/users:") // not the path block
}

// TODO(endpoint-review): top-level (info/servers/tags/security) slice tests once
// those blocks are wired, and a real-pipeline test with sources populated.
