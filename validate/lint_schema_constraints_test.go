package validate

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

const schemaConstraintsSpec = `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses: { "200": { description: ok } }
components:
  schemas:
    MultipleOfZero:   { type: integer, multipleOf: 0 }
    MultipleOfNeg:    { type: integer, multipleOf: -2 }
    BadRange:         { type: integer, minimum: 10, maximum: 5 }
    BadLength:        { type: string, minLength: 10, maxLength: 5 }
    BadItems:         { type: array, items: { type: string }, minItems: 5, maxItems: 2 }
    BadProps:         { type: object, minProperties: 5, maxProperties: 2 }
    EmptyEnum:        { type: string, enum: [] }
    EmptyAllOf:       { allOf: [] }
    ConstOutsideEnum: { type: string, enum: [x, y], const: z }
`

const goodSchemaConstraintsSpec = `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses: { "200": { description: ok } }
components:
  schemas:
    OkMultipleOf: { type: integer, multipleOf: 2 }
    OkRange:      { type: integer, minimum: 5, maximum: 10 }
    OkEqual:      { type: integer, minimum: 5, maximum: 5 }
    OkLength:     { type: string, minLength: 1, maxLength: 5 }
    OkItems:      { type: array, items: { type: string }, minItems: 1, maxItems: 5 }
    OkProps:      { type: object, minProperties: 1, maxProperties: 5 }
    OkEnum:       { type: string, enum: [x] }
    OkConst:      { type: string, enum: [x, y], const: x }
    OkAllOf:      { allOf: [ { type: object } ] }
    OnlyMin:      { type: integer, minimum: 10 }
    OnlyMax:      { type: integer, maximum: 5 }
`

func findingsByID(t *testing.T, spec string) map[string]int {
	t.Helper()
	counts := map[string]int{}
	for _, f := range lintSchemaConstraints(mustLoad(t, spec), "spec.yaml") {
		require.Equal(t, "spec.yaml", f.Source.File)
		require.NotEmpty(t, f.Fingerprint)
		require.NotEmpty(t, f.Section)
		counts[f.Id]++
	}
	return counts
}

// Each unsatisfiable or spec-violating combination is reported once.
func TestLintSchemaConstraints(t *testing.T) {
	got := findingsByID(t, schemaConstraintsSpec)

	require.Equal(t, 2, got[MultipleOfNotPositiveID], "zero and negative multipleOf")
	require.Equal(t, 1, got[MinimumExceedsMaximumID])
	require.Equal(t, 1, got[MinLengthExceedsMaxLengthID])
	require.Equal(t, 1, got[MinItemsExceedsMaxItemsID])
	require.Equal(t, 1, got[MinPropertiesExceedsMaxPropsID])
	require.Equal(t, 1, got[EnumEmptyID])
	require.Equal(t, 1, got[SubschemasEmptyID])
	require.Equal(t, 1, got[ConstNotInEnumID])
}

// A literal MUST violation is an error; an unsatisfiable combination the spec
// does not forbid outright is a warning.
func TestLintSchemaConstraints_Severities(t *testing.T) {
	levels := map[string]checker.Level{}
	for _, f := range lintSchemaConstraints(mustLoad(t, schemaConstraintsSpec), "spec.yaml") {
		levels[f.Id] = f.Level
	}
	require.Equal(t, checker.ERR, levels[MultipleOfNotPositiveID], "multipleOf > 0 is a MUST")
	require.Equal(t, checker.ERR, levels[SubschemasEmptyID], "non-empty allOf is a MUST")
	require.Equal(t, checker.WARN, levels[MinimumExceedsMaximumID])
	require.Equal(t, checker.WARN, levels[EnumEmptyID])
	require.Equal(t, checker.WARN, levels[ConstNotInEnumID])
}

// Satisfiable schemas produce nothing, including the equal-bounds edge case and
// a bound set on only one side.
func TestLintSchemaConstraints_Clean(t *testing.T) {
	require.Empty(t, lintSchemaConstraints(mustLoad(t, goodSchemaConstraintsSpec), "spec.yaml"))
}
