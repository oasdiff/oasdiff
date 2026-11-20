package validate

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// manyFindingsSpec puts enough findings in one spec for an arbitrary order to
// show: each readOnly property of a request-only schema is one finding, and
// the lint reads them from a map.
func manyFindingsSpec(properties int) string {
	var b strings.Builder
	b.WriteString(`
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'
      responses: { "200": { description: ok } }
components:
  schemas:
    Item:
      type: object
      properties:
`)
	for i := range properties {
		fmt.Fprintf(&b, "        p%02d:\n          type: string\n          readOnly: true\n", i)
	}
	return b.String()
}

func mustLoadWithOrigins(t *testing.T, data string) *openapi3.T {
	t.Helper()

	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	spec, err := loader.LoadFromData([]byte(data))
	require.NoError(t, err)
	return spec
}

// The same spec validates to the same findings in the same order every time.
// The lints walk maps, and so does kin when it reports its own errors, so
// without a sort the output moves between runs of one binary.
func TestValidate_SameOrderEveryRun(t *testing.T) {
	spec := mustLoadWithOrigins(t, manyFindingsSpec(12))

	first := Validate(spec, "spec.yaml")
	require.Len(t, first, 12)

	for range 20 {
		require.Equal(t, first, Validate(spec, "spec.yaml"))
	}
}

// The order is the source order, which is what a reader follows down the file.
func TestValidate_SortedBySourceLocation(t *testing.T) {
	findings := Validate(mustLoadWithOrigins(t, manyFindingsSpec(12)), "spec.yaml")

	lines := make([]int, 0, len(findings))
	for _, finding := range findings {
		lines = append(lines, finding.Source.Line)
	}
	require.True(t, slices.IsSorted(lines), "findings are not in source order: %v", lines)
}
