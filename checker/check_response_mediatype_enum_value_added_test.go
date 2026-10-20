package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// adding a value to a response body enum sends clients a value they may not
// handle, as for a response property
func TestResponseMediaTypeEnumValueAdded(t *testing.T) {
	changes := positionChanges(t, "response-body", enumSchema("a", "b"), enumSchema("a", "b", "c"))
	require.Len(t, changes, 1)
	require.Equal(t, checker.ResponseMediaTypeEnumValueAddedId, changes[0].GetId())
	require.Equal(t, checker.ERR, changes[0].GetLevel())
	localizer := checker.NewDefaultLocalizer()
	require.Equal(t, "added the new `c` enum value to the response schema `application/json` for the response status `200`", changes[0].GetUncolorizedText(localizer))
	require.Contains(t, changes[0].GetComment(localizer), "x-extensible-enum")
}
