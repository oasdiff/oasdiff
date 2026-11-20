//go:build windows

package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

// Windows terminals are not coloured, so ColorAuto renders plain even when
// the output is not piped.
func TestChangeText_PrettyNotPiped(t *testing.T) {
	piped := false
	save := checker.SetPipedOutput(&piped)
	defer checker.SetPipedOutput(save)

	require.Equal(t,
		"error\t[change_id] at revision.yaml\n\tin API GET /test\n\t\tThis is a breaking change.\n\t\tcomment",
		formatters.MultiLineErrorForTest(apiChange, MockLocalizer, checker.ColorAuto))

	require.Equal(t,
		"error\t[change_id]\n\tin security\n\t\tThis is a breaking change.\n\t\tcomment",
		formatters.MultiLineErrorForTest(securityChange, MockLocalizer, checker.ColorAuto))

	require.Equal(t,
		"error\t[change_id]\n\tin components/component\n\t\tThis is a breaking change.\n\t\tcomment",
		formatters.MultiLineErrorForTest(componentChange, MockLocalizer, checker.ColorAuto))
}
