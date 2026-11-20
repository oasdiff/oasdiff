//go:build unix

package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

// With colour, the level and the id are painted, and so is the element of the
// subject, which for an api change is the operation and path.
func TestChangeText_WithColor(t *testing.T) {
	require.Equal(t,
		"\x1b[31merror\x1b[0m at revision.yaml, in API \x1b[32mGET /test\x1b[0m This is a breaking change. [\x1b[33mchange_id\x1b[0m]. comment",
		formatters.SingleLineErrorForTest(apiChange, MockLocalizer, checker.ColorAlways))

	require.Equal(t,
		"\x1b[31merror\x1b[0m\t[\x1b[33mchange_id\x1b[0m]\n\tin security\n\t\tThis is a breaking change.\n\t\tcomment",
		formatters.MultiLineErrorForTest(securityChange, MockLocalizer, checker.ColorAlways))

	require.Equal(t,
		"\x1b[31merror\x1b[0m, in components/component This is a breaking change. [\x1b[33mchange_id\x1b[0m]. comment",
		formatters.SingleLineErrorForTest(componentChange, MockLocalizer, checker.ColorAlways))
}

// Output that is not piped is coloured under ColorAuto.
func TestChangeText_PrettyNotPiped(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)

	require.Equal(t,
		"\x1b[31merror\x1b[0m\t[\x1b[33mchange_id\x1b[0m] at revision.yaml\n\tin API \x1b[32mGET /test\x1b[0m\n\t\tThis is a breaking change.\n\t\tcomment",
		formatters.MultiLineErrorForTest(apiChange, MockLocalizer, checker.ColorAuto))
}
