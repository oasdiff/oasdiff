package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// removing the enum keyword from a response schema lets the server return values a client does not expect
func TestResponseMediaTypeBecameNotEnum(t *testing.T) {
	changes := checkChanges(t, checker.ResponseMediaTypeBecameNotEnumCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	requireSingleApiChange(t, checker.ApiChange{
		Id:          checker.ResponseMediaTypeBecameNotEnumId,
		Args:        []any{"text/plain", "200"},
		Comment:     checker.ResponseMediaTypeBecameNotEnumId + "-comment",
		Operation:   "PUT",
		Path:        "/mode",
		Source:      load.NewSource(enumRemovedEntirelyRevision),
		OperationId: "setMode",
	}, changes)
	require.Equal(t, checker.ERR, changes[0].GetLevel())
	require.Equal(t, "removed the enum constraint from the response schema `text/plain` for the response status `200`", changes[0].GetUncolorizedText(checker.NewDefaultLocalizer()))
}

func TestResponseMediaTypeBecameNotEnum_EnumAdded(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.ResponseMediaTypeBecameNotEnumCheck, enumRemovedEntirelyRevision, enumRemovedEntirelyBase))
}

// removing the enum keyword is one change, not one removed value per enum entry
func TestResponseMediaTypeEnumValueRemoved_EnumRemovedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.ResponseMediaTypeEnumValueRemovedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}
