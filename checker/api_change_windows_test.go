package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestApiChange_PrettyNotPipedWindows(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "error\t[change_id] at source\n\tin API GET /test\n\t\tThis is a breaking change.\n\t\tcomment", apiChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}
