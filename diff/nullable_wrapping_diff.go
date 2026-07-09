package diff

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// NullableWrappingDiff marks that a schema was made nullable by wrapping it in
// a oneOf with a bare null alternative: base X becomes
// oneOf: [{type: "null"}, X'] where X' is validation-equivalent to X. This is
// the common OpenAPI 3.1 idiom for making a $ref'd schema nullable, since a
// $ref cannot carry a type array.
//
// The wrap is equivalent to adding "null" to X's type set: the null branch
// cannot overlap X (the detector requires X to reject null), so oneOf's
// exactly-one rule behaves as anyOf here. Recognizing it keeps naive top-level
// comparisons (enum, pattern, type) from reading X's constraints as removed;
// they moved into the branch unchanged.
//
// Third member of the wrapping-recognition family, after ListOfTypesDiff
// (scalar type sets) and OneOfWrappingDiff (object alternatives, a breaking
// restructuring): this is the provably-safe wrapping shape those two don't
// classify. See oasdiff/oasdiff#1088.
type NullableWrappingDiff struct {
	// NullabilityAdded is true when the base schema was wrapped (the revision
	// accepts everything the base did, plus null).
	NullabilityAdded bool `json:"nullabilityAdded,omitempty" yaml:"nullabilityAdded,omitempty"`
}

// Empty indicates whether a change was found in this element.
func (diff *NullableWrappingDiff) Empty() bool {
	return diff == nil || !diff.NullabilityAdded
}

// getNullableWrappingDiff detects the nullable-wrapping pattern (revision
// wraps a schema equivalent to the base in oneOf: [{type: "null"}, base]) and
// returns nil when it doesn't apply. Only this direction is detected; the
// reverse (removing the wrap, i.e. removing nullability) keeps today's
// conservative reporting.
func getNullableWrappingDiff(config *Config, base, revision *openapi3.Schema) *NullableWrappingDiff {
	if base == nil || revision == nil {
		return nil
	}
	// The base must itself be free of compositions (so the equivalence below is
	// meaningful) and must reject null: wrapping an already-nullable schema
	// changes acceptance under oneOf's exactly-one rule (null would match both
	// branches and be rejected), so it is not a pure widening.
	if len(base.OneOf) > 0 || len(base.AnyOf) > 0 || len(base.AllOf) > 0 || base.Not != nil {
		return nil
	}
	if schemaAcceptsNull(base) {
		return nil
	}
	// The revision must be a bare wrapper: exactly a two-branch oneOf, nothing
	// constraining at the top level.
	if len(revision.OneOf) != 2 || !constrainsNothingBeyondOneOf(config, revision) {
		return nil
	}
	var payload *openapi3.SchemaRef
	nullBranches := 0
	for _, ref := range revision.OneOf {
		if isBareNullSchema(config, schemaValue(ref)) {
			nullBranches++
			continue
		}
		payload = ref
	}
	if nullBranches != 1 || payload == nil {
		return nil
	}
	if !SchemaRefsValidationEquivalent(config, &openapi3.SchemaRef{Value: base}, payload) {
		return nil
	}
	return &NullableWrappingDiff{NullabilityAdded: true}
}

// schemaAcceptsNull reports whether the schema accepts a null value via the
// nullable keyword (3.0) or a "null" entry in its type set (3.1).
func schemaAcceptsNull(schema *openapi3.Schema) bool {
	if schema.Nullable {
		return true
	}
	return slices.Contains(schema.Type.Slice(), "null")
}

// constrainsNothingBeyondOneOf reports whether the schema, with its oneOf
// cleared, is validation-equivalent to the empty schema (which accepts
// everything). Checking by equivalence rather than by a hand-maintained field
// list means a validation keyword this package doesn't anticipate makes the
// recognition decline (safe) instead of silently passing (unsound).
func constrainsNothingBeyondOneOf(config *Config, schema *openapi3.Schema) bool {
	bare := *schema
	bare.OneOf = nil
	return SchemaRefsValidationEquivalent(config, &openapi3.SchemaRef{Value: &bare}, emptySchemaRef())
}

// isBareNullSchema reports whether the schema accepts exactly null and nothing
// else: type ["null"], with everything beyond the type validation-equivalent
// to the empty schema (same rationale as constrainsNothingBeyondOneOf).
func isBareNullSchema(config *Config, schema *openapi3.Schema) bool {
	if schema == nil {
		return false
	}
	types := schema.Type.Slice()
	if len(types) != 1 || types[0] != "null" {
		return false
	}
	bare := *schema
	bare.Type = nil
	return SchemaRefsValidationEquivalent(config, &openapi3.SchemaRef{Value: &bare}, emptySchemaRef())
}

func emptySchemaRef() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{}}
}
