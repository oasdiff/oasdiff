package internal

import (
	"fmt"

	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

// nullFile is the path git passes when one side of a diff doesn't exist
// (file added on this commit, file deleted on this commit, or one side of
// a root-commit diff). It's the reliable signal — git's hash field varies
// here (40 zeros in the steady-state added/deleted case, "." for root
// commits — see git's diff.c) but the path is always "/dev/null".
const nullFile = "/dev/null"

func getGitDiffDriverCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "git-diff-driver path old-file old-hex old-mode new-file new-hex new-mode",
		Short: "Run as a git external diff driver to show OpenAPI changes inline in git log/diff",
		Long: `Run as a git external diff driver. Git invokes the driver with seven positional
arguments: the in-tree path, the old/new working-tree files, the old/new blob
hashes, and the old/new modes. oasdiff loads the two blobs directly via "git
cat-file" and prints the changelog between them.

Wire it up with two git config entries and a .gitattributes line:

    git config diff.oasdiff.command "oasdiff git-diff-driver"
    echo "openapi.yaml diff=oasdiff" >> .gitattributes

Then view OpenAPI changes inline with:

    git log --patch --ext-diff openapi.yaml

This subcommand is normally invoked by git, not by humans. It accepts the same
configuration as the changelog subcommand via the .oasdiff.yaml file in the
repository root.
`,
		Args: cobra.ExactArgs(7),
		RunE: runGitDiffDriver,
	}

	addCommonDiffFlags(&cmd)
	addCommonBreakingFlags(&cmd)
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), ""), "fail-on", "o", "ignored in git-diff-driver mode (exit code must stay 0 to keep the diff pipeline alive)")
	enumWithOptions(&cmd, newEnumValue(GetSupportedLevels(), LevelInfo), "level", "", "output errors with this level or higher")

	return &cmd
}

func runGitDiffDriver(cmd *cobra.Command, args []string) error {
	// Positional args follow git's external-diff protocol:
	//   args[0] = in-tree path        (e.g. "openapi.yaml")
	//   args[1] = old working file    ("/dev/null" when file didn't exist)
	//   args[2] = old blob hash       (40 zeros, or "." for root commits)
	//   args[3] = old mode            (unused)
	//   args[4] = new working file    ("/dev/null" when file no longer exists)
	//   args[5] = new blob hash
	//   args[6] = new mode            (unused)
	path := args[0]
	oldFile := args[1]
	oldHex := args[2]
	newFile := args[4]
	newHex := args[5]

	if oldFile == nullFile {
		fmt.Fprintf(cmd.OutOrStdout(), "Added %s\n", path)
		return nil
	}
	if newFile == nullFile {
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", path)
		return nil
	}
	if oldHex == newHex {
		// Mode-only change. Git's own diff machinery already prints the mode
		// transition; we have nothing meaningful to add.
		return nil
	}

	flags := NewFlags()
	if returnErr := RunViper(cmd, flags.getViper()); returnErr != nil {
		// External diff drivers must exit 0 to keep `git log --ext-diff` alive.
		// Surface the configuration error inline in the diff output instead of
		// aborting git's whole pipeline.
		fmt.Fprintf(cmd.OutOrStdout(), "oasdiff: configuration error: %s\n", returnErr)
		return nil
	}
	flags.setBase(load.NewSource(shortHex(oldHex) + ":" + path))
	flags.setRevision(load.NewSource(shortHex(newHex) + ":" + path))

	cmd.Root().SilenceUsage = true

	if _, returnErr := runChangelog(flags, cmd.OutOrStdout()); returnErr != nil {
		// Same exit-code-zero discipline as above. Print the error in-line so
		// the user sees it as part of git's diff output.
		fmt.Fprintf(cmd.OutOrStdout(), "oasdiff: %s\n", returnErr)
	}

	return nil
}

// shortHex abbreviates a 40-char git blob hash to its first 7 characters, matching
// git's own conventional short-hash length. Anything shorter than 8 is returned
// as-is (already short or empty).
func shortHex(hex string) string {
	const shortLen = 7
	if len(hex) <= shortLen {
		return hex
	}
	return hex[:shortLen]
}
