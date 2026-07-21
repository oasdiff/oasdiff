package validate

import (
	"fmt"
	"testing"

	"github.com/oasdiff/oasdiff/formatters"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func lintParamSpec(t *testing.T, spec string) formatters.Findings {
	t.Helper()
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	return lintAmbiguousParamSerialization(doc, "test.yaml")
}

func TestLintAmbiguousParamSerialization_TypeUnions(t *testing.T) {
	const specTemplate = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      parameters:
        - in: query
          name: token
          schema:
            type: [%s]
            items: { type: string }
      responses:
        "200": { description: ok }
`
	for _, tc := range []struct {
		types   string
		flagged bool
	}{
		{"array, integer", true},
		{"object, string", true},
		{"array, integer, \"null\"", true}, // null is ignored, the rest still mixes
		{"array, \"null\"", false},         // null carries no serialization
		{"string, \"null\"", false},
		{"string, integer", false}, // two scalars both serialize as a single value
		{"array", false},
	} {
		t.Run(tc.types, func(t *testing.T) {
			findings := lintParamSpec(t, fmt.Sprintf(specTemplate, tc.types))
			if !tc.flagged {
				require.Empty(t, findings)
				return
			}
			require.Len(t, findings, 1)
			f := findings[0]
			require.Equal(t, AmbiguousParameterSerializationID, f.Id)
			require.Equal(t, checker.WARN, f.Level)
			require.Equal(t, "/paths/~1a/get/parameters/0", f.Section)
			require.Contains(t, f.Text, `parameter "token"`)
			require.Equal(t, "test.yaml", f.Source.File)
			require.Equal(t, 10, f.Source.Line, "points at the schema's type field")
			require.NotEmpty(t, f.Fingerprint)
		})
	}
}

// A shared components parameter referenced from several operations is reported
// once, at the components definition.
func TestLintAmbiguousParamSerialization_SharedComponentsParam(t *testing.T) {
	const spec = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      parameters:
        - $ref: '#/components/parameters/Token'
      responses:
        "200": { description: ok }
  /b:
    get:
      parameters:
        - $ref: '#/components/parameters/Token'
      responses:
        "200": { description: ok }
components:
  parameters:
    Token:
      in: query
      name: token
      schema:
        type: [array, integer]
        items: { type: string }
`
	findings := lintParamSpec(t, spec)
	require.Len(t, findings, 1)
	require.Equal(t, "/components/parameters/Token", findings[0].Section)
}

// Parameters without a schema (content-based) and path-item-level parameters
// are handled; a 3.0 single-type parameter cannot express a union.
func TestLintAmbiguousParamSerialization_Shapes(t *testing.T) {
	const contentParam = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    parameters:
      - in: query
        name: filter
        content:
          application/json:
            schema: { type: object }
    get:
      responses:
        "200": { description: ok }
`
	require.Empty(t, lintParamSpec(t, contentParam), "content-based parameters have no style serialization")

	const pathItemParam = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    parameters:
      - in: query
        name: token
        schema:
          type: [object, boolean]
    get:
      responses:
        "200": { description: ok }
`
	findings := lintParamSpec(t, pathItemParam)
	require.Len(t, findings, 1)
	require.Equal(t, "/paths/~1a/parameters/0", findings[0].Section)
}

func TestLintAmbiguousParamSerialization_Webhooks(t *testing.T) {
	const spec = `openapi: 3.1.0
info: { title: t, version: "1" }
webhooks:
  newThing:
    post:
      parameters:
        - in: header
          name: X-Batch
          schema:
            type: [array, string]
            items: { type: string }
      responses:
        "200": { description: ok }
`
	findings := lintParamSpec(t, spec)
	require.Len(t, findings, 1)
	require.Equal(t, "/webhooks/newThing/post/parameters/0", findings[0].Section)
}

// The lint is wired into Validate alongside the kin findings.
func TestValidate_IncludesAmbiguousParamSerialization(t *testing.T) {
	const spec = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      parameters:
        - in: query
          name: token
          schema:
            type: [array, integer]
            items: { type: string }
      responses:
        "200": { description: ok }
`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	var ids []string
	for _, f := range Validate(doc, "test.yaml") {
		ids = append(ids, f.Id)
	}
	require.Contains(t, ids, AmbiguousParameterSerializationID)
}

// The walker also reaches parameters inside callback operations.
func TestLintAmbiguousParamSerialization_Callbacks(t *testing.T) {
	const spec = `openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /subscribe:
    post:
      callbacks:
        onEvent:
          '{$request.body#/url}':
            post:
              parameters:
                - in: query
                  name: attempt
                  schema:
                    type: [array, integer]
                    items: { type: string }
              responses:
                "200": { description: ok }
      responses:
        "201": { description: created }
`
	findings := lintParamSpec(t, spec)
	require.Len(t, findings, 1)
	require.Equal(t, "/paths/~1subscribe/post/callbacks/onEvent/{$request.body#~1url}/post/parameters/0", findings[0].Section)
}
