package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// restricting a response property to an enum narrows what the server returns
func TestResponsePropertyBecameEnum(t *testing.T) {
	changes := checkChanges(t, checker.ResponsePropertyBecameEnumCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase)
	requireApiChanges(t, []checker.ApiChange{
		{
			Id:          checker.ResponsePropertyBecameEnumId,
			Args:        []any{"status", "200"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyBase),
			OperationId: "createPet",
		},
		{
			Id:          checker.ResponsePropertyBecameEnumId,
			Args:        []any{"secret", "200"},
			Operation:   "POST",
			Path:        "/pets",
			Source:      load.NewSource(enumRemovedEntirelyBase),
			OperationId: "createPet",
		},
	}, changes)
	for _, c := range changes {
		require.Equal(t, checker.INFO, c.GetLevel())
	}
}

func TestResponsePropertyBecameEnum_EnumRemoved(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.ResponsePropertyBecameEnumCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}

// adding an enum keyword is one narrowing change, not one added value per enum entry
func TestResponsePropertyEnumValueAdded_EnumAddedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.ResponsePropertyEnumValueAddedCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}
