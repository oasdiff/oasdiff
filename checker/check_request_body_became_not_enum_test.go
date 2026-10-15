package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// removing the enum keyword from a request body accepts every value it accepted before
func TestRequestBodyBecameNotEnum(t *testing.T) {
	changes := checkChanges(t, checker.RequestBodyBecameNotEnumCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.RequestBodyBecameNotEnumId,
		Operation:   "PUT",
		Path:        "/mode",
		Source:      load.NewSource(enumRemovedEntirelyRevision),
		OperationId: "setMode",
	}, changes)
	require.Equal(t, checker.INFO, changes[0].GetLevel())
	require.Equal(t, "removed the enum constraint from the request body", changes[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestRequestBodyBecameNotEnum_EnumAdded(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestBodyBecameNotEnumCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}

// removing the enum keyword is one change, not one removed value per enum entry
func TestRequestBodyEnumValueRemoved_EnumRemovedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.RequestBodyEnumValueRemovedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}
