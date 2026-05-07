package internal

import (
	"io"
	"strconv"

	"github.com/oasdiff/oasdiff/build"
	"github.com/spf13/cobra"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {

	rootCmd := &cobra.Command{
		Use:   "oasdiff",
		Short: "OpenAPI specification diff",
	}

	rootCmd.SetArgs(args[1:])
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.Version = build.Version

	// --config is a persistent flag on the root command so every subcommand
	// inherits it. Lookup order is documented in internal/viper.go's
	// readConfFile: --config > OASDIFF_CONFIG env var > .oasdiff.* in cwd.
	rootCmd.PersistentFlags().String("config", "", "path to config file (overrides .oasdiff.* lookup; can also use the OASDIFF_CONFIG env var)")

	rootCmd.AddCommand(
		getDiffCmd(),
		getSummaryCmd(),
		getBreakingChangesCmd(),
		getChangelogCmd(),
		getFlattenCmd(),
		getChecksCmd(),
		getValidateCmd(),
	)

	return run(rootCmd)
}

func setReturnValue(cmd *cobra.Command, code int) {
	if cmd.Root().Annotations == nil {
		cmd.Root().Annotations = map[string]string{}
	}

	cmd.Root().Annotations["return"] = strconv.Itoa(code)
}

func getReturnValue(cmd *cobra.Command) int {
	if cmd.Root().Annotations == nil {
		return 0
	}

	codeStr := cmd.Root().Annotations["return"]
	if codeStr == "" {
		return 0
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		// this shouldn't happen
		return 0
	}

	return code
}

func run(cmd *cobra.Command) int {

	if err := cmd.Execute(); err != nil {
		if ret := getReturnValue(cmd); ret != 0 {
			return ret
		}
		return generalExecutionErr
	}

	return getReturnValue(cmd)
}
