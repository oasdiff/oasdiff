package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

func getBreakingFilesCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "breaking-files spec [spec...] --base ref [flags]",
		Short: "Display breaking changes in each spec against its own version in a git ref",
		Long: `Display breaking changes in each spec against its own version in a git ref.

Each argument is a spec file in the working tree, compared against the same path in
--base. A spec that isn't in the base ref is newly added and is skipped, since a new
API has no prior version to break. The command fails if any spec has changes at or
above --fail-on, so it can gate a commit or a CI job that already knows which files
changed. See .pre-commit-hooks.yaml for the pre-commit wiring.

Results are printed one spec at a time, so a structured --format emits one document
per spec rather than a single combined one.`,
		Args: getBreakingFilesArgs(),
		RunE: runBreakingFiles,
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetBreakingLevels(), ""), "fail-on", "o", "exit with return code 1 when output includes errors with this level or higher")
	cmd.PersistentFlags().String("base", "", "git ref holding the version to compare each spec against (e.g. origin/main)")

	// Composed mode merges the specs matching two globs into a single
	// comparison, which is the opposite of comparing each spec on its own.
	hideFlag(&cmd, "composed")

	return &cmd
}

// getBreakingFilesArgs requires --base and at least one spec, and rejects an
// argument that isn't a plain file path.
func getBreakingFilesArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if base, _ := cmd.Flags().GetString("base"); base == "" {
			return errors.New("please specify --base, the git ref holding the version to compare against (e.g. --base origin/main)")
		}
		if len(args) < 1 {
			return errors.New("please specify one or more spec files to compare against the base ref")
		}
		// Each argument is joined to the base ref as "<ref>:<path>", so it has to
		// be a plain file path. Stdin, a URL and a git revision have no version in
		// the ref to compare against, and joining them produces nonsense like
		// "HEAD:HEAD:openapi.yaml". Check what the argument is rather than listing
		// what it must not be, so a source type added later is covered.
		for _, arg := range args {
			if !load.NewSource(arg).IsFile() {
				return fmt.Errorf("%q is not a file path: every argument must be a spec file that also has a version in the base ref", arg)
			}
		}
		if composed, err := cmd.Flags().GetBool("composed"); err == nil && composed {
			return errors.New("--composed merges specs into one comparison, which breaking-files does not do; use 'oasdiff breaking --composed' instead")
		}

		return checkCommonFlags(cmd)
	}
}

func runBreakingFiles(cmd *cobra.Command, args []string) error {

	flags := NewFlags()
	if err := RunViper(cmd, flags.getViper()); err != nil {
		setReturnValue(cmd, err.Code)
		return err
	}

	// flags parsed successfully, so don't show usage on comparison errors
	cmd.Root().SilenceUsage = true

	base, _ := cmd.Flags().GetString("base")

	failed := false
	for _, file := range args {
		// Name each spec so a multi-file run says which result belongs to which
		// one; otherwise a bare "No changes detected" is ambiguous.
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "=== %s ===\n", file)

		baseSource := load.NewSource(base + ":" + file)
		baseSource.Fetch = flags.getFetch()
		flags.setBase(baseSource)

		revSource := load.NewSource(file)
		revSource.Fetch = flags.getFetch()
		flags.setRevision(revSource)

		failOn, err := runBreakingChanges(flags, cmd.OutOrStdout())
		if err != nil {
			// A spec that isn't in the base ref but is in the working tree is
			// newly added: it has no prior version, and a new API cannot contain
			// breaking changes. Skip it and keep checking the rest.
			//
			// The working-tree check is what separates a new spec from a mistyped
			// one. Without it a path in neither place looks new, and a typo would
			// pass the check it was meant to fail.
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

// fileExists reports whether path names an existing file in the working tree.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
