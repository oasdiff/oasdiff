package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

var securityChange = checker.SecurityChange{
	Id:      "change_id",
	Comment: "comment",
	Level:   checker.ERR,
	Args:    []any{1},
}

func TestSecurityChange(t *testing.T) {
	require.Equal(t, "security", securityChange.GetSection())
	require.Equal(t, "comment", securityChange.GetComment(MockLocalizer))
	require.Equal(t, "", securityChange.GetOperationId())
	require.Equal(t, []any{1}, securityChange.GetArgs())
}

func TestSecurityChange_MatchIgnore(t *testing.T) {
	require.True(t, securityChange.MatchIgnore("", "error, in security this is a breaking change. [change_id]. comment", MockLocalizer))
}
