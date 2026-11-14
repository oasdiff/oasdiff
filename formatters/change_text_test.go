package formatters_test

import (
	"testing"

	"github.com/TwiN/go-color"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

var apiChange = checker.ApiChange{
	Id:          "change_id",
	Args:        []any{},
	Comment:     "comment_id",
	Level:       checker.ERR,
	Operation:   "GET",
	OperationId: "123",
	Path:        "/test",
	CommonChange: checker.CommonChange{
		BaseSource:     &checker.Source{File: "base.yaml", Line: 10, Column: 5},
		RevisionSource: &checker.Source{File: "revision.yaml", Line: 12, Column: 7},
	},
}

var securityChange = checker.SecurityChange{
	Id:      "change_id",
	Comment: "comment",
	Level:   checker.ERR,
	Args:    []any{1},
}

var componentChange = checker.ComponentChange{
	Id:        "change_id",
	Comment:   "comment",
	Level:     checker.ERR,
	Component: "component",
}

// Each change type's single line and three-line forms, which differ only in
// what the change is about and whether it has a file to name.
func TestChangeText(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change checker.Change
		single string
		multi  string
	}{
		{
			name:   "api change",
			change: apiChange,
			single: "error at revision.yaml, in API GET /test This is a breaking change. [change_id]. comment",
			multi:  "error\t[change_id] at revision.yaml\n\tin API GET /test\n\t\tThis is a breaking change.\n\t\tcomment",
		},
		{
			name:   "security change",
			change: securityChange,
			single: "error, in security This is a breaking change. [change_id]. comment",
			multi:  "error\t[change_id]\n\tin security\n\t\tThis is a breaking change.\n\t\tcomment",
		},
		{
			name:   "component change",
			change: componentChange,
			single: "error, in components/component This is a breaking change. [change_id]. comment",
			multi:  "error\t[change_id]\n\tin components/component\n\t\tThis is a breaking change.\n\t\tcomment",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.single, formatters.SingleLineErrorForTest(tt.change, MockLocalizer, checker.ColorNever))
			require.Equal(t, tt.multi, formatters.MultiLineErrorForTest(tt.change, MockLocalizer, checker.ColorNever))
		})
	}
}

// A change with no comment ends after its text.
func TestChangeText_NoComment(t *testing.T) {
	change := apiChange
	change.Comment = ""

	require.Equal(t,
		"error\t[change_id] at revision.yaml\n\tin API GET /test\n\t\tThis is a breaking change.",
		formatters.MultiLineErrorForTest(change, MockLocalizer, checker.ColorNever))
}

// An info change names the file the version is in, or no file at all when the
// spec has none.
func TestChangeText_InfoChange(t *testing.T) {
	withSource := checker.InfoChange{
		Id:    "change_id",
		Level: checker.INFO,
	}.WithSources(nil, &checker.Source{File: "revision.yaml", Line: 4, Column: 3})

	require.Equal(t,
		"info\t[change_id] at revision.yaml\n\tin info\n\t\tThis is a breaking change.",
		formatters.MultiLineErrorForTest(withSource, MockLocalizer, checker.ColorNever))

	none := checker.InfoChange{Id: "change_id", Level: checker.INFO}
	require.Equal(t,
		"info\t[change_id]\n\tin info\n\t\tThis is a breaking change.",
		formatters.MultiLineErrorForTest(none, MockLocalizer, checker.ColorNever))
	require.Equal(t,
		"info, in info This is a breaking change. [change_id]. ",
		formatters.SingleLineErrorForTest(none, MockLocalizer, checker.ColorNever))
}

// A level reads as its name, coloured by severity only when colour is on.
func TestLevelText(t *testing.T) {
	for _, level := range []checker.Level{checker.ERR, checker.WARN, checker.INFO, checker.Level(4)} {
		require.Equal(t, level.String(), formatters.LevelTextForTest(level, checker.ColorNever))
		require.Equal(t, level.String(), formatters.LevelTextForTest(level, checker.ColorInvalid))
		require.Contains(t, formatters.LevelTextForTest(level, checker.ColorAlways), level.String())
		require.NotEqual(t, level.String(), formatters.LevelTextForTest(level, checker.ColorAlways),
			"a coloured level carries escape codes")
	}

	require.Equal(t, color.InRed("error"), formatters.LevelTextForTest(checker.ERR, checker.ColorAlways))
	require.Equal(t, color.InPurple("warning"), formatters.LevelTextForTest(checker.WARN, checker.ColorAlways))
	require.Equal(t, color.InCyan("info"), formatters.LevelTextForTest(checker.INFO, checker.ColorAlways))
	require.Equal(t, color.InGray("issue"), formatters.LevelTextForTest(checker.Level(4), checker.ColorAlways))
}
