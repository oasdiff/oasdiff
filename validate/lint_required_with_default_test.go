package validate

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

const requiredWithDefaultSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      parameters:
        - name: version        # required + default -> flag
          in: query
          required: true
          schema:
            type: string
            default: v1
        - name: page           # optional + default -> ok
          in: query
          required: false
          schema:
            type: integer
            default: 1
        - name: id             # required, no default -> ok
          in: query
          required: true
          schema:
            type: string
      responses: { "200": { description: ok } }
components:
  schemas:
    User:
      type: object
      properties:
        region:                # required + default -> flag
          type: string
          default: us-east-1
        name:                  # required, no default -> ok
          type: string
        note:                  # optional + default -> ok
          type: string
          default: ""
        address:               # nested object; its required prop + default -> flag (recursive)
          type: object
          properties:
            zip:
              type: string
              default: "00000"
          required: [zip]
      required: [region, name, address]
`

const cleanRequiredSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
      responses: { "200": { description: ok } }
components:
  schemas:
    User:
      type: object
      properties:
        region:
          type: string
          default: us-east-1
      required: []
`

// WARN on a required parameter with a default and a required property with a
// default; leave optional-with-default and required-without-default alone.
func TestLintRequiredWithDefault(t *testing.T) {
	findings := lintRequiredWithDefault(mustLoad(t, requiredWithDefaultSpec), "spec.yaml")
	require.Len(t, findings, 3) // required param, top-level required property, and the nested one

	texts := map[string]bool{}
	sections := map[string]bool{}
	for _, f := range findings {
		require.Equal(t, RequiredWithDefaultID, f.Id)
		require.Equal(t, checker.WARN, f.Level)
		require.Equal(t, "spec.yaml", f.Source.File)
		require.NotEmpty(t, f.Fingerprint)
		require.NotEmpty(t, f.Section)
		texts[f.Text] = true
		sections[f.Section] = true
	}
	require.True(t, texts[`required parameter "version" has a default value, which is never used because a required parameter must always be sent`])
	require.True(t, texts[`required property "region" has a default value, which is never used because a required property must always be present`])
	// WalkSchemas is recursive, so a required-with-default nested inside another
	// object is caught too, at its exact JSON pointer.
	require.True(t, sections["/components/schemas/User/properties/address/properties/zip"])
}

// A spec with only optional-with-default and required-without-default produces
// no findings.
func TestLintRequiredWithDefault_Clean(t *testing.T) {
	require.Empty(t, lintRequiredWithDefault(mustLoad(t, cleanRequiredSpec), "spec.yaml"))
}
