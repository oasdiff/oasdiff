//go:build unix

package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestSecurityChange_PrettyNotPipedUnix(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)
	require.Equal(t, "\x1b[31merror\x1b[0m\t[\x1b[33mchange_id\x1b[0m]\n\tin security\n\t\tThis is a breaking change.\n\t\tcomment", securityChange.MultiLineError(MockLocalizer, checker.ColorAuto))
}

func TestSecurityChange_SingleLineError_WithColor(t *testing.T) {
	require.Equal(t, "\x1b[31merror\x1b[0m, in security This is a breaking change. [\x1b[33mchange_id\x1b[0m]. comment", securityChange.SingleLineError(MockLocalizer, checker.ColorAlways))
}
