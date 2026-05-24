//go:build unix

package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestApiChange_PrettyNotPipedUnix(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "\x1b[31merror\x1b[0m\t[\x1b[33mchange_id\x1b[0m] at source\n\tin API \x1b[32mGET\x1b[0m \x1b[32m/test\x1b[0m\n\t\tThis is a breaking change.\n\t\tcomment", apiChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}

func TestApiChange_SingleLineError_WithColor(t *testing.T) {
	require.Equal(t, "\x1b[31merror\x1b[0m at source, in API \x1b[32mGET\x1b[0m \x1b[32m/test\x1b[0m This is a breaking change. [\x1b[33mchange_id\x1b[0m]. comment", apiChange.SingleLineError(MockLocalizer, checker.ColorAlways))
}
