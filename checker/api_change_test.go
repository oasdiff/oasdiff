package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

var apiChange = checker.ApiChange{
	Id:              "change_id",
	Args:            []any{},
	Comment:         "comment_id",
	Level:           checker.ERR,
	Operation:       "GET",
	OperationId:     "123",
	Path:            "/test",
	Source:          load.NewSource("source"),
	SourceFile:      "sourceFile",
	SourceLine:      1,
	SourceLineEnd:   2,
	SourceColumn:    3,
	SourceColumnEnd: 4,
}

func TestApiChange(t *testing.T) {
	require.Equal(t, "paths", apiChange.GetSection())
	require.Equal(t, "change_id", apiChange.GetId())
	require.Equal(t, "comment", apiChange.GetComment(MockLocalizer))
	require.Equal(t, checker.ERR, apiChange.GetLevel())
	require.Equal(t, "GET", apiChange.GetOperation())
	require.Equal(t, "123", apiChange.GetOperationId())
	require.Equal(t, "/test", apiChange.GetPath())
	require.Equal(t, "source", apiChange.GetSource())
	require.Equal(t, "sourceFile", apiChange.GetSourceFile())
	require.Equal(t, 1, apiChange.GetSourceLine())
	require.Equal(t, 2, apiChange.GetSourceLineEnd())
	require.Equal(t, 3, apiChange.GetSourceColumn())
	require.Equal(t, 4, apiChange.GetSourceColumnEnd())
	require.Equal(t, "error at source, in API GET /test This is a breaking change. [change_id]. comment", apiChange.SingleLineError(MockLocalizer, checker.ColorNever))
}

func MockLocalizer(originalKey string, args ...any) string {
	switch originalKey {
	case "change_id":
		return "This is a breaking change."
	case "comment_id":
		return "comment"
	default:
		return originalKey
	}

}

func TestApiChange_MatchIgnore(t *testing.T) {
	require.True(t, apiChange.MatchIgnore("/test", "error at source, in api get /test this is a breaking change. [change_id]. comment", MockLocalizer))
}

func TestApiChange_MultiLineError(t *testing.T) {
	require.Equal(t, "error\t[change_id] at source\t\n\tin API GET /test\n\t\tThis is a breaking change.\n\t\tcomment", apiChange.MultiLineError(MockLocalizer, checker.ColorNever))
}

func TestApiChange_MultiLineError_NoComment(t *testing.T) {
	apiChangeNoComment := apiChange
	apiChangeNoComment.Comment = ""

	require.Equal(t, "error\t[change_id] at source\t\n\tin API GET /test\n\t\tThis is a breaking change.", apiChangeNoComment.MultiLineError(MockLocalizer, checker.ColorNever))
}

func TestApiChange_SourceFile(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("spec.yaml")

	require.Equal(t, "spec.yaml", apiChangeSourceFile.GetSourceFile())
}

func TestApiChange_SourceUrl(t *testing.T) {
	apiChangeSourceFile := apiChange
	apiChangeSourceFile.SourceFile = ""
	apiChangeSourceFile.Source = load.NewSource("http://google.com/spec.yaml")

	require.Equal(t, "", apiChangeSourceFile.GetSourceFile())
}
