package internal

import (
	"errors"
	"io"

	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/cobra"
)

const specHelp = `
Base and revision can be a path to a file, a URL, a git ref (e.g. main:openapi.yaml), or '-' to read standard input.
In 'composed' mode, base and revision can be a glob and oasdiff will compare matching endpoints between the two sets of files.`

func getParseArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("please specify base and revision arguments as a path to a file, a glob (in composed mode), a URL, a git ref (e.g. main:openapi.yaml), or '-' to read standard input")
		}
		if len(args) > 2 {
			return errors.New("invalid arguments after base and revision")
		}
		if err := checkStdinWithComposed(cmd, args); err != nil {
			return err
		}
		if err := checkOpenWithComposed(cmd); err != nil {
			return err
		}
		if err := checkReviewFlagsRequireOpen(cmd); err != nil {
			return err
		}
		if err := checkColor(cmd); err != nil {
			return err
		}

		return nil
	}
}

type runner func(flags *Flags, stdout io.Writer) (bool, *ReturnError)

func getRun(runner runner) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {

		flags := NewFlags()

		if err := RunViper(cmd, flags.getViper()); err != nil {
			setReturnValue(cmd, err.Code)
			return err
		}

		if len(args) > 0 {
			base := load.NewSource(args[0])
			base.Fetch = flags.getFetch()
			flags.setBase(base)
		}

		if len(args) > 1 {
			revision := load.NewSource(args[1])
			revision.Fetch = flags.getFetch()
			flags.setRevision(revision)
		}

		// by now flags have been parsed successfully so we don't need to show usage on any errors
		cmd.Root().SilenceUsage = true

		failEmpty, err := runner(flags, cmd.OutOrStdout())
		if err != nil {
			setReturnValue(cmd, err.Code)
			return err
		}

		if failEmpty {
			setReturnValue(cmd, 1)
		}

		return nil
	}
}

func checkColor(cmd *cobra.Command) error {

	if colorPassed := cmd.Flags().Changed("color"); !colorPassed {
		return nil
	}

	if format, _ := cmd.Flags().GetString("format"); format == "text" || format == "singleline" {
		return nil
	}

	return errors.New(`--color flag is only relevant with 'text' or 'singleline' formats`)
}

func checkOpenWithComposed(cmd *cobra.Command) error {

	// --open exists only on breaking and changelog; diff and summary share
	// getParseArgs but don't define it.
	if cmd.Flags().Lookup("open") == nil {
		return nil
	}

	open, err := cmd.Flags().GetBool("open")
	if err != nil {
		return errors.New("failed to get open flag")
	}

	if !open {
		return nil
	}

	composed, err := cmd.Flags().GetBool("composed")
	if err != nil {
		return errors.New("failed to get composed flag")
	}

	if composed {
		// --open builds a side-by-side review of exactly two specs; composed
		// mode (-c) diffs a glob of many files, which the review can't
		// represent.
		return errors.New("--open cannot be used with composed mode (-c): the side-by-side review compares exactly two specs")
	}

	return nil
}

func checkReviewFlagsRequireOpen(cmd *cobra.Command) error {

	// Only breaking/changelog define these (via addOpenFlags); skip elsewhere.
	if cmd.Flags().Lookup("review-token") == nil {
		return nil
	}

	// --open is registered alongside the review flags; the Lookup is defensive
	// in case it ever isn't.
	open := false
	if cmd.Flags().Lookup("open") != nil {
		var err error
		if open, err = cmd.Flags().GetBool("open"); err != nil {
			return errors.New("failed to get open flag")
		}
	}
	if open {
		return nil
	}

	if cmd.Flags().Changed("review-token") || cmd.Flags().Changed("review-meta") {
		return errors.New("--review-token and --review-meta require --open")
	}

	return nil
}

func checkStdinWithComposed(cmd *cobra.Command, args []string) error {

	composed, err := cmd.Flags().GetBool("composed")
	if err != nil {
		return errors.New("failed to get composed flag")
	}

	if !composed {
		return nil
	}

	if args[0] == "-" || args[1] == "-" {
		return errors.New("can't read from stdin in composed mode")
	}

	return nil
}
