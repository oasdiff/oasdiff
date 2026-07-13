package diff_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// An OAS 3.1 Reference Object metadata override ($ref sibling `description`)
// changes the effective document: adding one must surface in the diff as a
// description change on the referenced parameter.
func TestDiff_RefMetadataOverride31(t *testing.T) {
	t.Skip("3.1 Reference Object metadata overrides land upstream in kin-openapi PR #1215; unskip at the kin bump that includes it")
	const base = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      parameters:
        - $ref: '#/components/parameters/P'
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
	const revision = `openapi: 3.1.0
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
	loadSide := func(data string) *load.SpecInfo {
		si, err := load.NewSpecInfoFromData(openapi3.NewLoader(), []byte(data), "spec.yaml")
		require.NoError(t, err)
		return si
	}

	d, err := diff.Get(diff.NewConfig(), loadSide(base).Spec, loadSide(revision).Spec)
	require.NoError(t, err)
	require.False(t, d.Empty(), "the override changes the effective document")

	paramsDiff := d.PathsDiff.Modified["/a"].OperationsDiff.Modified["GET"].ParametersDiff
	require.NotNil(t, paramsDiff)
	descDiff := paramsDiff.Modified["query"]["q"].DescriptionDiff
	require.NotNil(t, descDiff)
	require.Equal(t, "original", descDiff.From)
	require.Equal(t, "overridden", descDiff.To)
}
