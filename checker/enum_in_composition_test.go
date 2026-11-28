package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func enumInCompositionChanges(t *testing.T, check checker.BackwardCompatibilityCheck) checker.Changes {
	t.Helper()

	return enumInCompositionChangesOf(t, "enum_in_composition", check)
}

func enumInCompositionChangesOf(t *testing.T, fixture string, check checker.BackwardCompatibilityCheck) checker.Changes {
	t.Helper()

	s1, err := open("../data/checker/" + fixture + "_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/" + fixture + "_revision.yaml")
	require.NoError(t, err)

	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)

	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(check), d, osm, checker.INFO)
}

func changeIds(changes checker.Changes) []string {
	ids := make([]string, 0, len(changes))
	for _, change := range changes {
		ids = append(ids, change.GetId())
	}

	return ids
}

// a value added to an enum in an anyOf branch next to a branch that accepts
// every string was already valid, so only the oneOf enum's new value is reported
func TestResponsePropertyEnumValueAddedInComposition(t *testing.T) {
	changes := enumInCompositionChanges(t, checker.ResponsePropertyEnumValueAddedCheck)

	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyEnumValueAddedId,
		Args:        []any{"medium", "size/oneOf[subschema #1]", "200"},
		Comment:     checker.ResponsePropertyEnumValueAddedId + "-comment",
		Operation:   "POST",
		Path:        "/pets",
		OperationId: "createPet",
	}, changes)
}

// a value removed from an enum in an anyOf branch stays valid through the
// branch that accepts every string, null included through a branch whose
// type list has it
func TestResponsePropertyEnumValueRemovedInComposition(t *testing.T) {
	require.Empty(t, enumInCompositionChanges(t, checker.ResponseParameterEnumValueRemovedCheck))
}

// in a request, the anyOf enum's removed value is still accepted and its
// added value already was, while the oneOf enum's added value is reported
func TestRequestPropertyEnumValueUpdatedInComposition(t *testing.T) {
	changes := enumInCompositionChanges(t, checker.RequestPropertyEnumValueUpdatedCheck)

	require.Equal(t, []string{checker.RequestPropertyEnumValueAddedId}, changeIds(changes))
	require.Equal(t, []any{"medium", "size/oneOf[subschema #1]"}, changes[0].GetArgs())
}

// an enum under items or additionalProperties restricts a different value
// than the schema holding it, an allOf branch's enum applies to every value,
// and a const sibling accepts only its own value, so each removal is reported
func TestResponsePropertyEnumValueRemovedOutsideOpenComposition(t *testing.T) {
	changes := enumInCompositionChangesOf(t, "enum_in_composition_reported", checker.ResponseParameterEnumValueRemovedCheck)

	require.ElementsMatch(t, [][]any{
		{"brown", "tags/items", "200"},
		{"brown", "traits/additionalProperties", "200"},
		{"bird", "kind/anyOf[subschema #1]", "200"},
		{"pug", "breed/allOf[subschema #2]", "200"},
	}, changeArgs(changes))
}

func TestRequestPropertyEnumValueRemovedOutsideOpenComposition(t *testing.T) {
	changes := enumInCompositionChangesOf(t, "enum_in_composition_reported", checker.RequestPropertyEnumValueUpdatedCheck)

	require.ElementsMatch(t, [][]any{
		{"brown", "tags/items"},
		{"brown", "traits/additionalProperties"},
		{"bird", "kind/anyOf[subschema #1]"},
		{"pug", "breed/allOf[subschema #2]"},
	}, changeArgs(changes))
}

// OpenAPI 3.0 reads nullable on the sibling branch, so null removed from the
// enum stays valid
func TestPropertyEnumValueRemovedInComposition30(t *testing.T) {
	require.Empty(t, enumInCompositionChangesOf(t, "enum_in_composition_30", checker.ResponseParameterEnumValueRemovedCheck))
	require.Empty(t, enumInCompositionChangesOf(t, "enum_in_composition_30", checker.RequestPropertyEnumValueUpdatedCheck))
}

func changeArgs(changes checker.Changes) [][]any {
	args := make([][]any, 0, len(changes))
	for _, change := range changes {
		args = append(args, change.GetArgs())
	}

	return args
}
