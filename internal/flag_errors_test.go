package internal

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// invalidValueErrorFor parses the given flag with the given raw value and
// returns the *pflag.InvalidValueError that pflag emits when the raw value
// fails the flag's Set() method.
func invalidValueErrorFor(t *testing.T, registerFlag func(fs *pflag.FlagSet), name, rawValue string) error {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	registerFlag(fs)
	return fs.Parse([]string{"--" + name + "=" + rawValue})
}

func TestFriendlyFlagError_BoolHint(t *testing.T) {
	err := invalidValueErrorFor(t, func(fs *pflag.FlagSet) { fs.Bool("flatten-params", false, "") }, "flatten-params", "x")
	require.Error(t, err)
	out := friendlyFlagError(&cobra.Command{}, err)
	require.EqualError(t, out, `invalid argument "x" for "--flatten-params" flag: must be true or false`)
}

func TestFriendlyFlagError_IntHint(t *testing.T) {
	err := invalidValueErrorFor(t, func(fs *pflag.FlagSet) { fs.Int("max-circular-dep", 5, "") }, "max-circular-dep", "foo")
	require.Error(t, err)
	out := friendlyFlagError(&cobra.Command{}, err)
	require.EqualError(t, out, `invalid argument "foo" for "--max-circular-dep" flag: must be an integer`)
}

func TestFriendlyFlagError_DurationHint(t *testing.T) {
	err := invalidValueErrorFor(t, func(fs *pflag.FlagSet) { fs.Duration("timeout", 0, "") }, "timeout", "soon")
	require.Error(t, err)
	out := friendlyFlagError(&cobra.Command{}, err)
	require.EqualError(t, out, `invalid argument "soon" for "--timeout" flag: must be a duration like 30s or 5m`)
}

func TestFriendlyFlagError_UnknownTypeUnchanged(t *testing.T) {
	// String flags accept anything, so we can't trigger an InvalidValueError
	// for them. Instead, hand friendlyFlagError an unrelated error and verify
	// it passes through.
	original := errors.New("unrelated failure")
	out := friendlyFlagError(&cobra.Command{}, original)
	require.Same(t, original, out)
}
