package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

var componentChange = checker.ComponentChange{
	Id:        "change_id",
	Comment:   "comment",
	Level:     checker.ERR,
	Component: "component",
}

func TestComponentChange(t *testing.T) {
	require.Equal(t, "components", componentChange.GetSection())
	require.Equal(t, "comment", componentChange.GetComment(MockLocalizer))
	require.Equal(t, "", componentChange.GetOperationId())
}

func TestComponentChange_MatchIgnore(t *testing.T) {
	require.True(t, componentChange.MatchIgnore("", "error, in components/component this is a breaking change. [change_id]. comment", MockLocalizer))
}
