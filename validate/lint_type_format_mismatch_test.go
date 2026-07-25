package validate

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

const typeFormatMismatchSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      parameters:
        - name: since          # string format on an integer -> flag
          in: query
          schema:
            type: integer
            format: date-time
      responses: { "200": { description: ok } }
components:
  schemas:
    Thing:
      type: object
      properties:
        size:                  # number format on an integer -> flag
          type: integer
          format: double
        created:               # correct pairing -> ok
          type: string
          format: date-time
        count:                 # correct pairing -> ok
          type: integer
          format: int64
        big:                   # bigint is recognized for integer -> ok
          type: integer
          format: bigint
        ratio:                 # correct pairing -> ok
          type: number
          format: float
        id:                    # standard JSON Schema string format -> ok
          type: string
          format: uuid
        custom:                # unrecognized format -> allowed by the spec, ok
          type: string
          format: my-internal-format
        untyped:               # no type to compare against -> ok
          format: uuid
        nested:
          type: object
          properties:
            when:              # nested mismatch -> flag (WalkSchemas recurses)
              type: integer
              format: date
`

const cleanTypeFormatSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses: { "200": { description: ok } }
components:
  schemas:
    Thing:
      type: object
      properties:
        created:
          type: string
          format: date-time
        count:
          type: integer
          format: int32
        custom:
          type: string
          format: something-bespoke
`

// A 3.1 nullable type array is compared by its non-null type: a string format
// on ["string","null"] is correct, on ["integer","null"] it is a mismatch.
const nullableTypeFormatSpec = `
openapi: 3.1.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses: { "200": { description: ok } }
components:
  schemas:
    Thing:
      type: object
      properties:
        ok:
          type: [string, "null"]
          format: uuid
        bad:
          type: [integer, "null"]
          format: uuid
`

// formatOwner is built by inverting formatsByType, which silently drops a
// format claimed by two types. Assert the invariant it relies on: every format
// appears under exactly one type, and every entry survives the inversion.
func TestFormatOwner_OneTypePerFormat(t *testing.T) {
	seen := map[string]string{}
	for declaredType, formats := range formatsByType {
		for _, format := range formats {
			if other, dup := seen[format]; dup {
				t.Errorf("format %q is claimed by both %q and %q; formatOwner can only record one", format, other, declaredType)
			}
			seen[format] = declaredType
			require.Equal(t, declaredType, formatOwner[format], "format %q lost its owner in the inversion", format)
		}
	}
	require.Len(t, formatOwner, len(seen))
}

// WARN on a known format that belongs to a different type; stay silent for
// correct pairings, unrecognized formats, and schemas with no declared type.
func TestLintTypeFormatMismatch(t *testing.T) {
	findings := lintTypeFormatMismatch(mustLoad(t, typeFormatMismatchSpec), "spec.yaml")
	require.Len(t, findings, 3) // the parameter, size, and the nested when

	texts := map[string]bool{}
	sections := map[string]bool{}
	for _, f := range findings {
		require.Equal(t, TypeFormatMismatchID, f.Id)
		require.Equal(t, checker.WARN, f.Level)
		require.Equal(t, "spec.yaml", f.Source.File)
		require.NotEmpty(t, f.Fingerprint)
		require.NotEmpty(t, f.Section)
		texts[f.Text] = true
		sections[f.Section] = true
	}
	require.True(t, texts[`format "date-time" is defined for type "string", not "integer", so it is ignored`])
	require.True(t, texts[`format "double" is defined for type "number", not "integer", so it is ignored`])
	require.True(t, sections["/components/schemas/Thing/properties/nested/properties/when"])
}

// A spec whose formats all match their type, or are custom, produces no findings.
func TestLintTypeFormatMismatch_Clean(t *testing.T) {
	require.Empty(t, lintTypeFormatMismatch(mustLoad(t, cleanTypeFormatSpec), "spec.yaml"))
}

// The "null" entry of a 3.1 nullable type array is ignored when matching.
func TestLintTypeFormatMismatch_NullableTypeArray(t *testing.T) {
	findings := lintTypeFormatMismatch(mustLoad(t, nullableTypeFormatSpec), "spec.yaml")
	require.Len(t, findings, 1)
	require.Equal(t, "/components/schemas/Thing/properties/bad", findings[0].Section)
	require.Equal(t, `format "uuid" is defined for type "string", not "integer", so it is ignored`, findings[0].Text)
}
