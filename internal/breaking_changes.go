package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

const breakingBaseHelp = `

With --base, the positional arguments are the changed spec files (not a base/revision pair)
and each one is compared against its own version in the given git ref. This is the shape a
pre-commit hook needs, where the tool receives the list of changed files: see .pre-commit-hooks.yaml.`

func getBreakingChangesCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "breaking base revision [flags]",
		Short: "Display breaking changes",
		Long:  "Display breaking changes between base and revision specs." + specHelp + breakingBaseHelp,
		Args:  getBreakingArgs(),
		RunE:  runBreaking,
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetBreakingLevels(), ""), "fail-on", "o", "exit with return code 1 when output includes errors with this level or higher")
	addOpenFlags(&cmd, "breaking changes")
	cmd.Flags().String("base", "", "compare each changed spec file against this git ref (e.g. origin/main); positional arguments become the changed files, one comparison per file")

	return &cmd
}

// getBreakingArgs validates the positional arguments for the breaking command.
// Without --base it keeps the standard base+revision pair (getParseArgs). With
// --base the positionals are the changed files, so at least one is required and
// the base/revision-pair rules (exactly two, composed globs) don't apply.
func getBreakingArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		base, _ := cmd.Flags().GetString("base")
		if base == "" {
			return getParseArgs()(cmd, args)
		}

		if len(args) < 1 {
			return errors.New("with --base, specify one or more spec files to compare against the base version")
		}
		// The positionals are working-tree files to compare against the base
		// ref, so neither stdin nor a glob has a version in that ref to compare
		// against. Reject them here rather than let them fail further in.
		if slices.Contains(args, "-") {
			return errors.New("can't read from stdin with --base: the positional arguments are files that also exist in the base ref")
		}
		if composed, err := cmd.Flags().GetBool("composed"); err == nil && composed {
			return errors.New("can't use --composed with --base: each file is compared against its own version in the base ref")
		}

		return checkCommonFlags(cmd)
	}
}

func runBreaking(cmd *cobra.Command, args []string) error {
	base, _ := cmd.Flags().GetString("base")
	if base == "" {
		return getRun(runBreakingChanges)(cmd, args)
	}
	return runBreakingBase(cmd, base, args)
}

// runBreakingBase runs the breaking-change check for every changed file against
// its version in the base git ref, printing each result and failing the command
// if any file has changes meeting the --fail-on threshold. This is the pre-commit
// entry point: pre-commit passes the changed files as positional arguments and
// only needs a non-zero exit to block the commit.
func runBreakingBase(cmd *cobra.Command, base string, files []string) error {

	flags := NewFlags()
	if err := RunViper(cmd, flags.getViper()); err != nil {
		setReturnValue(cmd, err.Code)
		return err
	}

	// flags parsed successfully, so don't show usage on comparison errors
	cmd.Root().SilenceUsage = true

	failed := false
	for _, file := range files {
		// Label each file so a multi-file run says which result belongs to which
		// spec (otherwise a bare "No changes detected" is ambiguous).
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "=== %s ===\n", file)

		// Same "<ref>:<path>" git-revision syntax the git-diff driver uses; the
		// revision is the file in the working tree.
		baseSource := load.NewSource(base + ":" + file)
		baseSource.Fetch = flags.getFetch()
		flags.setBase(baseSource)

		revSource := load.NewSource(file)
		revSource.Fetch = flags.getFetch()
		flags.setRevision(revSource)

		failOn, err := runBreakingChanges(flags, cmd.OutOrStdout())
		if err != nil {
			// A file that isn't in the base ref but is in the working tree is
			// newly added: it has no prior version, and a new API cannot contain
			// breaking changes. Skip it and keep checking the remaining files
			// instead of aborting the whole run.
			//
			// The working-tree check is what separates a new file from a
			// mistyped one. Without it a path in neither place looks new, and a
			// typo would pass the check it was supposed to fail.
			if fileExists(file) && load.IsPathMissingInRef(base, file) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "new file, not in base ref, skipped")
				continue
			}
			setReturnValue(cmd, err.Code)
			return err
		}
		if failOn {
			failed = true
		}
	}

	if failed {
		setReturnValue(cmd, 1)
	}

	return nil
}

func runBreakingChanges(flags *Flags, stdout io.Writer) (bool, *ReturnError) {
	return getChangelog(flags, stdout, checker.WARN, true)
}

// fileExists reports whether path names an existing file in the working tree.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
