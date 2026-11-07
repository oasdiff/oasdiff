package allof_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/flatten/allof"
	"github.com/stretchr/testify/require"
)

// The merged schema is assembled from the branches, so it has no origin of its
// own. It keeps the origin of the schema that held the allOf, which is what a
// consumer reporting a location on the merged schema has to point at.
func TestMergeSpec_KeepsTheOriginOfTheSchemaThatHeldTheAllOf(t *testing.T) {
	const spec = `
openapi: 3.0.3
info:
  title: origins
  version: "1"
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        name:
          allOf:
            - type: string
              minLength: 1
            - maxLength: 10
`
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	name := doc.Components.Schemas["User"].Value.Properties["name"].Value
	require.NotNil(t, name.Origin, "the loader tracks the origin of the schema that holds the allOf")
	line, column := name.Origin.Key.Line, name.Origin.Key.Column

	merged, err := allof.MergeSpec(doc)
	require.NoError(t, err)

	name = merged.Components.Schemas["User"].Value.Properties["name"].Value
	require.Empty(t, name.AllOf, "the subschemas are merged")
	require.Equal(t, uint64(1), name.MinLength)
	require.Equal(t, uint64(10), *name.MaxLength)

	require.NotNil(t, name.Origin, "the merged schema keeps an origin")
	require.Equal(t, line, name.Origin.Key.Line)
	require.Equal(t, column, name.Origin.Key.Column)
}
