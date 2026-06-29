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

func TestStructuralKey(t *testing.T) {
	for _, tc := range []struct {
		name           string
		op, path       string
		wantKey, wantT string
	}{
		{"operation", "POST", "/users", "POST /users", "POST /users"},
		{"path level", "", "/users", "/users", "/users"},
		{"component", "", "components/schemas/User", "components/schemas/User", "components/schemas/User"},
		{"fallback", "", "", otherChangesKey, "Other changes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, title := structuralKey(checker.ApiChange{Operation: tc.op, Path: tc.path})
			require.Equal(t, tc.wantKey, key)
			require.Equal(t, tc.wantT, title)
		})
	}
}

// TODO(endpoint-review): add component-schema and top-level (info/servers)
// end-to-end slice tests once those mapping cases are wired.
