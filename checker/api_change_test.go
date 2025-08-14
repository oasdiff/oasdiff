package checker_test

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// TestMain sets up global test configuration for the checker package
func TestMain(m *testing.M) {
	// Enable origin tracking in openapi3 for all tests
	openapi3.IncludeOrigin = true

	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}

var apiChange = checker.ApiChange{
	Id:          "change_id",
	Args:        []any{},
	Comment:     "comment_id",
	Level:       checker.ERR,
	Operation:   "GET",
	OperationId: "123",
	Path:        "/test",
	Source:      load.NewSource("source"),
	CommonChange: checker.CommonChange{
		BaseSource:     checker.NewSource("base.yaml", 10, 5),
		RevisionSource: checker.NewSource("revision.yaml", 12, 7),
	},
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

	// Test new BaseSource and RevisionSource methods
	baseSource := apiChange.GetBaseSource()
	require.Equal(t, "base.yaml", baseSource.File)
	require.Equal(t, 10, baseSource.Line)
	require.Equal(t, 5, baseSource.Column)
	require.NotEmpty(t, baseSource)

	revisionSource := apiChange.GetRevisionSource()
	require.Equal(t, "revision.yaml", revisionSource.File)
	require.Equal(t, 12, revisionSource.Line)
	require.Equal(t, 7, revisionSource.Column)
	require.NotEmpty(t, revisionSource)

	require.Equal(t, "error at source, in API GET /test This is a breaking change. [change_id]. comment", apiChange.SingleLineError(MockLocalizer, checker.ColorNever))
}

func MockLocalizer(originalKey string, args ...interface{}) string {
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
