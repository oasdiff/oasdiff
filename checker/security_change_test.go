package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

var securityChange = checker.SecurityChange{
	Id:              "change_id",
	Comment:         "comment",
	Level:           checker.ERR,
	Args:            []any{1},
	SourceFile:      "sourceFile",
	SourceLine:      1,
	SourceLineEnd:   2,
	SourceColumn:    3,
	SourceColumnEnd: 4,
}

func TestSecurityChange(t *testing.T) {
	require.Equal(t, "security", securityChange.GetSection())
	require.Equal(t, "comment", securityChange.GetComment(MockLocalizer))
	require.Equal(t, "", securityChange.GetOperationId())
	require.Equal(t, "", securityChange.GetSource())
	require.Equal(t, []any{1}, securityChange.GetArgs())
	require.Equal(t, "sourceFile", securityChange.GetSourceFile())
	require.Equal(t, 1, securityChange.GetSourceLine())
	require.Equal(t, 2, securityChange.GetSourceLineEnd())
	require.Equal(t, 3, securityChange.GetSourceColumn())
	require.Equal(t, 4, securityChange.GetSourceColumnEnd())
	require.Equal(t, "error, in security This is a breaking change. [change_id]. comment", securityChange.SingleLineError(MockLocalizer, checker.ColorNever))
}

func TestSecurityChange_MatchIgnore(t *testing.T) {
	require.True(t, securityChange.MatchIgnore("", "error, in security this is a breaking change. [change_id]. comment", MockLocalizer))
}

func TestSecurityChange_SingleLineError(t *testing.T) {
	require.Equal(t, "error, in security This is a breaking change. [change_id]. comment", securityChange.SingleLineError(MockLocalizer, checker.ColorNever))
}

func TestSecurityChange_MultiLineError_NoColor(t *testing.T) {
	require.Equal(t, "error\t[change_id]\n\tin security\n\t\tThis is a breaking change.\n\t\tcomment", securityChange.MultiLineError(MockLocalizer, checker.ColorNever))
}
