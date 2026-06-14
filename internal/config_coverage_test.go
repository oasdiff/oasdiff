package internal

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// flagsNotInConfigFile lists persistent flags that are intentionally NOT
// settable from a config file, so the drift guard below skips them:
//   - open: an interactive one-shot action (encrypt the comparison, upload it,
//     and open a browser). Persisting it in a committed config would upload and
//     open a review on every run, so it is command-line only.
//   - config: the path to the config file itself.
//
// Hidden (deprecated) flags are skipped separately via flag.Hidden, so they
// don't need to be listed here.
var flagsNotInConfigFile = map[string]bool{
	"open":   true,
	"config": true,
}

// Test_ConfigFileCoversAllFlags guards against config drift.
//
// validateViperConfig unmarshals the config file into the Config struct with
// UnmarshalExact, which rejects any key not present in the struct. So a
// persistent flag missing from Config doesn't just go unread, it makes a config
// file that sets that flag fail to load entirely (every oasdiff command exits
// with a config error). This test asserts the inverse: every visible,
// non-excluded persistent flag on a config-loading command has a matching
// Config field, so users can put any real flag in .oasdiff.* without surprise.
func Test_ConfigFileCoversAllFlags(t *testing.T) {
	configKeys := map[string]bool{}
	for field := range reflect.TypeFor[Config]().Fields() {
		if tag := field.Tag.Get("mapstructure"); tag != "" {
			configKeys[tag] = true
		}
	}

	// Every command whose RunE is getRun(...) loads a config file via RunViper.
	commands := []*cobra.Command{
		getDiffCmd(),
		getSummaryCmd(),
		getBreakingChangesCmd(),
		getChangelogCmd(),
		getChecksCmd(),
		getFlattenCmd(),
		getUpgradeCmd(),
		getValidateCmd(),
	}

	for _, cmd := range commands {
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden || flagsNotInConfigFile[f.Name] {
				return
			}
			require.Truef(t, configKeys[f.Name],
				"flag %q on the %q command is settable on the command line but missing from the Config struct in viper.go. "+
					"Add a mapstructure field for it, or add it to flagsNotInConfigFile if it must not be config-settable.",
				f.Name, cmd.Name())
		})
	}
}
