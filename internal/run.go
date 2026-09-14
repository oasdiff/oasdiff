package internal

import (
	"io"
	"strconv"

	"github.com/oasdiff/oasdiff/build"
	"github.com/spf13/cobra"
)

const (
	groupCompare    = "compare"
	groupSingleSpec = "single-spec"
	groupGit        = "git"
	groupReference  = "reference"
)

func inGroup(cmd *cobra.Command, groupID string) *cobra.Command {
	cmd.GroupID = groupID
	return cmd
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {

	rootCmd := &cobra.Command{
		Use:   "oasdiff",
		Short: "OpenAPI specification diff",
		// Warn (on stderr) about any deprecated hidden flag the user set, for
		// every subcommand. Runs after flag parsing, before the command body.
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			warnDeprecatedFlags(cmd)
		},
	}

	rootCmd.SetArgs(args[1:])
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetFlagErrorFunc(friendlyFlagError)
	rootCmd.Version = build.Version

	// --config is a persistent flag on the root command so every subcommand
	// inherits it. Lookup order is documented in internal/viper.go's
	// readConfFile: --config > OASDIFF_CONFIG env var > .oasdiff.* in cwd.
	rootCmd.PersistentFlags().String("config", "", "path to config file (overrides .oasdiff.* lookup; can also use the OASDIFF_CONFIG env var)")

	rootCmd.AddGroup(
		&cobra.Group{ID: groupCompare, Title: "Compare two specs:"},
		&cobra.Group{ID: groupSingleSpec, Title: "Process a single spec:"},
		&cobra.Group{ID: groupGit, Title: "Git integration:"},
		&cobra.Group{ID: groupReference, Title: "Reference:"},
	)

	rootCmd.AddCommand(
		inGroup(getBreakingChangesCmd(), groupCompare),
		inGroup(getChangelogCmd(), groupCompare),
		inGroup(getDiffCmd(), groupCompare),
		inGroup(getSummaryCmd(), groupCompare),
		inGroup(getValidateCmd(), groupSingleSpec),
		inGroup(getUpgradeCmd(), groupSingleSpec),
		inGroup(getFlattenCmd(), groupSingleSpec),
		inGroup(getBreakingFilesCmd(), groupGit),
		inGroup(getGitDiffDriverCmd(), groupGit),
		inGroup(getChecksCmd(), groupReference),
		inGroup(getSchemaCmd(), groupReference),
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
