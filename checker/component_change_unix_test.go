//go:build unix

package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestComponentChange_SetPipedOutput_NotPiped_Unix(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "\x1b[31merror\x1b[0m\t[\x1b[33mchange_id\x1b[0m] \t\n\tin components/component\n\t\tThis is a breaking change.\n\t\tcomment", componentChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}

func TestComponentChange_SetPipedOutput_NilPiped_Unix(t *testing.T) {
	save := checker.SetPipedOutput(nil)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "error\t[change_id] \t\n\tin components/component\n\t\tThis is a breaking change.\n\t\tcomment", componentChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}

func TestComponentChange_SingleLineError_WithColor(t *testing.T) {
	require.Equal(t, "\x1b[31merror\x1b[0m, in components/component This is a breaking change. [\x1b[33mchange_id\x1b[0m]. comment", componentChange.SingleLineError(MockLocalizer, checker.ColorAlways))
}
