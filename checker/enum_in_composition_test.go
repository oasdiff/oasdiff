package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func enumInCompositionChanges(t *testing.T, check checker.BackwardCompatibilityCheck) checker.Changes {
	t.Helper()

	s1, err := open("../data/checker/enum_in_composition_base.yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/enum_in_composition_revision.yaml")
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
// branch that accepts every string
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
