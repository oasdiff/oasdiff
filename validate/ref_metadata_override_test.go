package validate

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// OAS 3.1 allows a $ref sibling `description` on parameter references (a
// Reference Object metadata override); OAS 3.0 does not. The two tests pin the
// version split.

const refSiblingDescriptionSpec = `openapi: %s
info: { title: t, version: "1" }
paths:
  /a:
    get:
      parameters:
        - $ref: '#/components/parameters/P'
          description: overridden
      responses:
        "200": { description: ok }
components:
  parameters:
    P:
      name: q
      in: query
      description: original
      schema: { type: string }
`

func findingIDsForVersion(t *testing.T, version string) []string {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromData(fmt.Appendf(nil, refSiblingDescriptionSpec, version))
	require.NoError(t, err)
	var ids []string
	for _, f := range Validate(doc, "test.yaml") {
		ids = append(ids, f.Id)
	}
	return ids
}

// In 3.0, metadata beside $ref is not part of the Reference Object: flagged.
func TestRefSiblingDescription_30_Flagged(t *testing.T) {
	require.Contains(t, findingIDsForVersion(t, "3.0.0"), "extra-sibling-fields")
}

// In 3.1, `description` beside a parameter $ref is a legal metadata override:
// no finding.
func TestRefSiblingDescription_31_Allowed(t *testing.T) {
	t.Skip("3.1 Reference Object metadata overrides land upstream in kin-openapi PR #1215; unskip at the kin bump that includes it")
	require.NotContains(t, findingIDsForVersion(t, "3.1.0"), "extra-sibling-fields")
}
