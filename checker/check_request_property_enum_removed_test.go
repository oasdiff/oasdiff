package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// removing the enum keyword from a request property accepts every value it accepted before
func TestRequestPropertyEnumRemoved(t *testing.T) {
	changes := checkChanges(t, checker.RequestPropertyEnumRemovedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.RequestPropertyEnumRemovedId,
			Args:        []any{"status"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyRevision),
			OperationId: "createPet",
		},
		{
			Id:          checker.RequestPropertyEnumRemovedId,
			Args:        []any{"kind"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyRevision),
			OperationId: "createPet",
		},
	}, changes)
	texts := []string{}
	for _, c := range changes {
		require.Equal(t, checker.INFO, c.GetLevel())
		texts = append(texts, c.GetUncolorizedText(checker.NewDefaultLocalizer()))
	}
	require.Contains(t, texts, "removed the enum constraint from the request property `status`")
}

// adding an enum keyword to a request property is reported by request-property-became-enum, not by this check
func TestRequestPropertyEnumRemoved_EnumAdded(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertyEnumRemovedCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}

// removing the enum keyword is one change, not one removed value per enum entry
func TestRequestPropertyEnumValueUpdated_EnumRemovedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertyEnumValueUpdatedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}

// adding an enum keyword is reported by request-property-became-enum, not as added values
func TestRequestPropertyEnumValueUpdated_EnumAddedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertyEnumValueUpdatedCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}

// an enum that moved into a oneOf branch may still restrict the values, so it is not reported as removed
func TestRequestPropertyEnumRemoved_MovedIntoOneOf(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestPropertyEnumRemovedCheck, "../data/checker/nullable_wrap_narrowed_base.yaml", "../data/checker/nullable_wrap_narrowed_revision.yaml"))
}
