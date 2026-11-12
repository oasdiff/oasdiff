package internal

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/versioning"
	"github.com/stretchr/testify/require"
)

// Every path that sets --fail-on validates the value first, the flag through
// its enum and the config file through its loader, so failOnVersionBump can
// only meet an invalid one if a future path forgets to. It fails the flags
// rather than reading the value as NONE and passing the run.
func TestFailOnVersionBump_InvalidFailOn(t *testing.T) {
	flags := NewFlags()
	flags.v.Set("fail-on", "nonsense")

	failEmpty, returnErr := failOnVersionBump(flags, checker.Changes{}, versioning.Verdict{Status: versioning.Satisfied})

	require.False(t, failEmpty)
	require.NotNil(t, returnErr)
	require.Equal(t, 101, returnErr.Code)
	require.ErrorContains(t, returnErr, `invalid fail-on value "nonsense"`)
}
