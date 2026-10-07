package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// removing the enum keyword from a request parameter, or from a property of one, accepts every value accepted before
func TestRequestParameterEnumRemoved(t *testing.T) {
	changes := checkChanges(t, checker.RequestParameterEnumRemovedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestParameterEnumRemovedId,
			Args:        []any{"query", "status"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyRevision),
			OperationId: "listPets",
		},
		{
			Id:          checker.RequestParameterPropertyEnumRemovedId,
			Args:        []any{"origin", "query", "filter"},
			Operation:   "GET",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyRevision),
			OperationId: "listPets",
		},
	}, changes)
	for _, c := range changes {
		require.Equal(t, checker.INFO, c.GetLevel())
	}
	require.Equal(t, "removed the enum constraint from the `query` request parameter `status`", requireChange(t, changes, checker.RequestParameterEnumRemovedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "removed the enum constraint from the property `origin` of the `query` request parameter `filter`", requireChange(t, changes, checker.RequestParameterPropertyEnumRemovedId).GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestRequestParameterEnumRemoved_EnumAdded(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestParameterEnumRemovedCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}

// removing the enum keyword is one change, not one removed value per enum entry
func TestRequestParameterEnumValueUpdated_EnumRemovedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestParameterEnumValueUpdatedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}

// adding an enum keyword to a parameter is reported by request-parameter-became-enum, not as added values;
// a parameter property has no became-enum rule outside headers, so its added values are still listed
func TestRequestParameterEnumValueUpdated_EnumAddedEntirely(t *testing.T) {
	changes := checkChanges(t, checker.RequestParameterEnumValueUpdatedCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase)
	require.Len(t, changes, 2)
	for _, c := range changes {
		require.Equal(t, checker.RequestParameterPropertyEnumValueAddedId, c.GetId())
	}
}
