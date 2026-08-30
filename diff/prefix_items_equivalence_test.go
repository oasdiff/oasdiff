package diff

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func stringRef() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
}

func integerRef() *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}}
}

func arraySchema(items *openapi3.SchemaRef, prefix ...*openapi3.SchemaRef) *openapi3.Schema {
	return &openapi3.Schema{Type: &openapi3.Types{"array"}, Items: items, PrefixItems: prefix}
}

// The schema governing array element i is prefixItems[i] where the list
// reaches, otherwise items. Adding an entry equal to the items schema makes
// explicit what items already governed: element 0 must be a string on both
// sides, and elements past the prefix are governed by the same items schema.
// Every array valid before is valid after and vice versa, so the change is an
// editorial restatement of the document with no contract effect.
func TestPrefixItemsEquivalent_EntryRestatesItems(t *testing.T) {
	base := arraySchema(stringRef())
	revision := arraySchema(stringRef(), stringRef())

	require.True(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// An entry that differs from the items schema re-governs its position:
// element 0 was validated by items ({type: string}) and is now validated by
// the entry ({type: integer}). ["a"] loses validity and [5] gains it, so
// neither accepted set contains the other, and ordering a differing pair is
// subsumption, which the diff does not decide. The helper answers only
// changed or unchanged; the direction stays undetermined.
func TestPrefixItemsEquivalent_EntryDiffersFromItems(t *testing.T) {
	base := arraySchema(stringRef())
	revision := arraySchema(stringRef(), integerRef())

	require.False(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// With neither a prefix entry nor an items schema, a position is governed by
// the empty schema and accepts anything. An entry with a type constrains a
// position that accepted every value, so the contract changed: [5] was a
// valid array and is rejected once element 0 must be a string. This is the
// regime where adding an entry narrows; under items: false the same edit
// widens (it admits a longer array), which is why the direction is a
// property of the items schema and not of the edit.
func TestPrefixItemsEquivalent_EntryConstrainsOpenPosition(t *testing.T) {
	base := arraySchema(nil)
	revision := arraySchema(nil, stringRef())

	require.False(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// The no-op mirror of the constrained case: an entry that accepts everything,
// at a position that accepted everything. The document gained an entry and no
// array changes validity in either direction. Together with the previous test
// this pins that the helper classifies by the entry's contract, not by the
// act of adding: same base, same syntactic edit, opposite verdicts decided by
// the entry's content alone.
func TestPrefixItemsEquivalent_EmptyEntryOnOpenPosition(t *testing.T) {
	base := arraySchema(nil)
	revision := arraySchema(nil, &openapi3.SchemaRef{Value: &openapi3.Schema{}})

	require.True(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// Identical lists govern every position identically.
func TestPrefixItemsEquivalent_IdenticalLists(t *testing.T) {
	base := arraySchema(stringRef(), integerRef())
	revision := arraySchema(stringRef(), integerRef())

	require.True(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// prefixItems is positional, so the comparison pairs index i with index i.
// Swapping the last two entries changes which schema governs positions 1 and
// 2 even though the lists hold the same schemas as a set.
func TestPrefixItemsEquivalent_Reorder(t *testing.T) {
	base := arraySchema(stringRef(), stringRef(), integerRef())
	revision := arraySchema(stringRef(), integerRef(), stringRef())

	require.False(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// The position comparison delegates to SchemaRefsValidationEquivalent, so
// whatever it considers contract-free is inherited here: a description on an
// otherwise identical entry does not change which arrays validate.
func TestPrefixItemsEquivalent_AnnotationOnly(t *testing.T) {
	annotated := integerRef()
	annotated.Value.Description = "the count"
	base := arraySchema(stringRef(), integerRef())
	revision := arraySchema(stringRef(), annotated)

	require.True(t, PrefixItemsValidationEquivalent(NewConfig(), base, revision))
}

// A nil schema has no positions to compare, so equivalence is never claimed.
func TestPrefixItemsEquivalent_NilSchema(t *testing.T) {
	require.False(t, PrefixItemsValidationEquivalent(NewConfig(), nil, arraySchema(stringRef())))
	require.False(t, PrefixItemsValidationEquivalent(NewConfig(), arraySchema(stringRef()), nil))
}
