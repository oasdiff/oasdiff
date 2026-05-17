package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestComponentChange_PrettyNotPipedWindows(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "error\t[change_id]\n\tin components/component\n\t\tThis is a breaking change.\n\t\tcomment", componentChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}
