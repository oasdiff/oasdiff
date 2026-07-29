package formatters

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/TwiN/go-color"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

type TEXTFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newTEXTFormatter(l checker.Localizer) TEXTFormatter {
	return TEXTFormatter{
		Localizer: l,
	}
}

func (f TEXTFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return []byte(report.GetTextReportAsString(diff)), nil
}

func (f TEXTFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	if len(changes) == 0 {
		if opts.DiffEmpty {
			_, _ = fmt.Fprint(result, "No changes detected")
		} else if opts.IsBreaking {
			_, _ = fmt.Fprint(result, "No breaking changes to report, but the specs are different.\nRun 'oasdiff diff' to see structural differences.")
		} else {
			_, _ = fmt.Fprint(result, "No changes to report, but the specs are different.\nRun 'oasdiff diff' to see structural differences.")
		}
		return result.Bytes(), nil
	}

	_, _ = fmt.Fprint(result, getChangelogTitle(changes, f.Localizer, opts.ColorMode))

	for _, c := range changes {
		_, _ = fmt.Fprintf(result, "%s\n\n", c.MultiLineError(f.Localizer, opts.ColorMode))
	}

	return result.Bytes(), nil
}

// RenderChecks prints one row per check. Columns follow the same rule as
// Check's omitempty json tags: a column that is empty for every row is left out
// rather than printed as a blank column under a header promising data. That
// keeps an IDs-only listing (`oasdiff checks validate`, where a rule's text and
// severity are only known at runtime) to a plain list of ids, while the
// changelog listing, which populates both, is unchanged.
func (f TEXTFormatter) RenderChecks(checks Checks, opts RenderOpts) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	withDescription := slices.ContainsFunc(checks, func(c Check) bool { return c.Description != "" })
	withLevel := slices.ContainsFunc(checks, func(c Check) bool { return c.Level != "" })

	row := func(id, description, level string) string {
		cells := []string{id}
		if withDescription {
			cells = append(cells, description)
		}
		if withLevel {
			cells = append(cells, level)
		}
		return strings.Join(cells, "\t")
	}

	w := tabwriter.NewWriter(result, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, row("ID", "DESCRIPTION", "LEVEL"))
	for _, check := range checks {
		_, _ = fmt.Fprintln(w, row(check.Id, f.Localizer(check.Description), check.Level))
	}
	_ = w.Flush()

	return result.Bytes(), nil
}

// RenderValidate emits a summary line ("N findings: ...") followed by one
// changelog-style block per finding. Each block is:
//
//	error	[<rule-id>] at <file:line:column>
//		<text>
//
// Findings with operation context add an "in API METHOD /path" line and
// indent the message one level deeper. Level is colorized and the rule ID
// rendered yellow when --color is on, matching changelog / breaking.
func (f TEXTFormatter) RenderValidate(findings Findings, opts RenderOpts) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	// Mirror RenderChangelog's empty case ("No changes detected") rather than
	// printing a "0 findings" summary, so the sibling commands read alike.
	if len(findings) == 0 {
		_, _ = fmt.Fprint(result, "No findings detected")
		return result.Bytes(), nil
	}

	count := findings.GetLevelCount()
	// Color the severity labels in the summary line, matching the changelog
	// command's title (getChangelogTitle).
	_, _ = fmt.Fprintf(result, "%d findings: %d %s, %d %s, %d %s\n",
		len(findings),
		count[checker.ERR], checker.ERR.StringCond(opts.ColorMode),
		count[checker.WARN], checker.WARN.StringCond(opts.ColorMode),
		count[checker.INFO], checker.INFO.StringCond(opts.ColorMode),
	)

	useColor := checker.IsColorEnabled(opts.ColorMode)
	for _, finding := range findings {
		loc := finding.Source.File
		if finding.Source.Line > 0 {
			loc = fmt.Sprintf("%s:%d:%d", finding.Source.File, finding.Source.Line, finding.Source.Column)
		}
		id := finding.Id
		if useColor {
			id = color.InYellow(finding.Id)
		}
		_, _ = fmt.Fprintf(result, "%s\t[%s] at %s\n", finding.Level.StringCond(opts.ColorMode), id, loc)

		msgIndent := "\t"
		if finding.Operation != "" && finding.Path != "" {
			operation, path := finding.Operation, finding.Path
			if useColor {
				// Color the endpoint green, matching ApiChange.MultiLineError.
				operation, path = color.InGreen(operation), color.InGreen(path)
			}
			_, _ = fmt.Fprintf(result, "\tin API %s %s\n", operation, path)
			msgIndent = "\t\t"
		}
		_, _ = fmt.Fprintf(result, "%s%s\n\n", msgIndent, indentContinuation(finding.Text, msgIndent))
	}

	return result.Bytes(), nil
}

func (f TEXTFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputChangelog, OutputChecks, OutputValidate}
}
