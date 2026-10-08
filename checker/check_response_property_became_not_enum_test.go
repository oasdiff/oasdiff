package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// removing the enum keyword from a response property lets the server return values a client does not expect
func TestResponsePropertyBecameNotEnum(t *testing.T) {
	changes := checkChanges(t, checker.ResponsePropertyBecameNotEnumCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	require.Len(t, changes, 2)

	status := changes[0]
	if status.(checker.ApiChange).Args[0] != "status" {
		status = changes[1]
	}
	requireApiChange(t, checker.ApiChange{
		Id:          checker.ResponsePropertyBecameNotEnumId,
		Args:        []any{"status", "200"},
		Comment:     checker.ResponsePropertyBecameNotEnumId + "-comment",
		Operation:   "POST",
		Path:        "/pets",
		Source:      load.NewSource(enumRemovedEntirelyRevision),
		OperationId: "createPet",
	}, status)
	require.Equal(t, checker.ERR, status.GetLevel())
	require.Equal(t, "removed the enum constraint from the `status` response property for the response status `200`", status.GetUncolorizedText(checker.NewDefaultLocalizer()))
	require.Equal(t, "The server may now return any value of the property's type, including values the previous enum excluded, so a client written against it may not handle the response.", status.GetComment(checker.NewDefaultLocalizer()))
}

// a write-only property never appears in responses, so removing its enum cannot break a client
func TestResponseWriteOnlyPropertyBecameNotEnum(t *testing.T) {
	changes := checkChanges(t, checker.ResponsePropertyBecameNotEnumCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision)
	for _, c := range changes {
		if c.(checker.ApiChange).Args[0] == "secret" {
			require.Equal(t, checker.INFO, c.GetLevel())
			return
		}
	}
	require.Fail(t, "expected a change for the write-only property")
}

// removing the enum keyword is one change, not one removed value per enum entry
func TestResponsePropertyEnumValueRemoved_EnumRemovedEntirely(t *testing.T) {
	require.Empty(t, checkChanges(t, checker.ResponseParameterEnumValueRemovedCheck, enumRemovedEntirelyBase, enumRemovedEntirelyRevision))
}
