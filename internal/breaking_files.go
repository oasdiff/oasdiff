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
changed, and it is what the oasdiff pre-commit hook runs. See
https://github.com/oasdiff/oasdiff/blob/main/docs/BREAKING-FILES.md

Results are printed one spec at a time, so a structured --format, or a --template,
is rendered once per spec rather than producing a single combined document.`,
		Args: getBreakingFilesArgs(),
		RunE: runBreakingFiles,
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetBreakingLevels(), ""), "fail-on", "o", "exit with return code 1 when output includes errors with this level or higher")
	cmd.PersistentFlags().String("base", "", "git ref holding the version to compare each spec against (e.g. origin/main)")

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
		// be a plain file path. Stdin, a URL and a git revision are not paths, and
		// joining them produces nonsense like "HEAD:HEAD:openapi.yaml". Whether the
		// path is in the ref is a separate question, answered per spec during the
		// run, since one that is not is a new file. Check what the argument is
		// rather than listing what it must not be, so a source type added later is
		// covered.
		for _, arg := range args {
			if !load.NewSource(arg).IsFile() {
				return fmt.Errorf("%q is not a file path: every argument must be a spec file in the working tree", arg)
			}
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
			// A spec in the working tree but not in the base ref is newly added,
			// and a new API cannot contain breaking changes, so skip it and keep
			// checking the rest. Requiring it in the working tree is what tells a
			// new spec from a mistyped path, which is absent from the ref too.
			if fileExists(file) && errors.Is(err, load.ErrPathNotInRef) {
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
