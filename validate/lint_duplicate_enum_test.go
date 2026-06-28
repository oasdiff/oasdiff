package validate

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func mustLoad(t *testing.T, data string) *openapi3.T {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromData([]byte(data))
	require.NoError(t, err)
	return spec
}

const dupEnumSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /x:
    get:
      parameters:
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive, active]
      responses: { "200": { description: ok } }
components:
  schemas:
    Color:
      type: string
      enum: [red, green, red, blue, green]
`

const cleanEnumSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Color:
      type: string
      enum: [red, green, blue]
`

// WARN on duplicate enum values, both in a component schema and an inline
// parameter schema; report each duplicated value once. (#980)
func TestLintDuplicateEnums(t *testing.T) {
	findings := lintDuplicateEnums(mustLoad(t, dupEnumSpec), "spec.yaml")
	require.Len(t, findings, 2)

	byDups := map[string][]string{}
	for _, f := range findings {
		require.Equal(t, DuplicateEnumValueID, f.Id)
		require.Equal(t, checker.WARN, f.Level)
		require.Equal(t, "spec.yaml", f.Source.File)
		require.NotEmpty(t, f.Fingerprint)
		byDups[f.Section] = nil
	}
	// One finding from the component schema, one from the inline parameter.
	// WalkSchemas reports each schema's exact RFC 6901 JSON Pointer as the section.
	require.Contains(t, byDups, "/components/schemas/Color")
	require.Contains(t, byDups, "/paths/~1x/get/parameters/0/schema")
}

// A duplicated value is reported once even when it repeats more than twice;
// distinct duplicates are all reported.
func TestDuplicateEnumValues(t *testing.T) {
	require.ElementsMatch(t, []string{"red", "green"},
		duplicateEnumValues([]any{"red", "green", "red", "blue", "green", "red"}))
	require.Empty(t, duplicateEnumValues([]any{"red", "green", "blue"}))
	// "1" (string) and 1 (number) are distinct, not a duplicate.
	require.Empty(t, duplicateEnumValues([]any{"1", 1}))
}

// No duplicate-enum finding for a clean spec.
func TestLintDuplicateEnums_Clean(t *testing.T) {
	require.Empty(t, lintDuplicateEnums(mustLoad(t, cleanEnumSpec), "spec.yaml"))
}

// End-to-end through Validate: the lint runs alongside kin validation, so a
// spec that is otherwise valid still surfaces the duplicate-enum WARN.
func TestValidate_SurfacesDuplicateEnum(t *testing.T) {
	findings := Validate(mustLoad(t, dupEnumSpec), "spec.yaml")
	var got int
	for _, f := range findings {
		if f.Id == DuplicateEnumValueID {
			got++
		}
	}
	require.Equal(t, 2, got)
}
