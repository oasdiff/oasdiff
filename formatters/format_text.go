package formatters

import (
	"bytes"
	"fmt"
	"text/tabwriter"

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

func (f TEXTFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, _, _ string) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	if len(changes) > 0 {
		_, _ = fmt.Fprint(result, getChangelogTitle(changes, f.Localizer, opts.ColorMode))
	}

	for _, c := range changes {
		_, _ = fmt.Fprintf(result, "%s\n\n", c.MultiLineError(f.Localizer, opts.ColorMode))
	}

	return result.Bytes(), nil
}

func (f TEXTFormatter) RenderChecks(checks Checks, opts RenderOpts) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	w := tabwriter.NewWriter(result, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tDESCRIPTION\tLEVEL")
	for _, check := range checks {
		_, _ = fmt.Fprintln(w, check.Id+"\t"+f.Localizer(check.Description)+"\t"+check.Level)
	}
	_ = w.Flush()

	return result.Bytes(), nil
}

func (f TEXTFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputChangelog, OutputChecks}
}
