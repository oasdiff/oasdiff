package formatters

import (
	"bytes"
	"fmt"

	"github.com/oasdiff/oasdiff/checker"
)

type SingleLineFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newSingleLineFormatter(l checker.Localizer) SingleLineFormatter {
	return SingleLineFormatter{
		Localizer: l,
	}
}

func (f SingleLineFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, _, _ string) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	if len(changes) > 0 {
		_, _ = fmt.Fprint(result, getChangelogTitle(changes, f.Localizer, opts.ColorMode))
	}

	for _, c := range changes {
		_, _ = fmt.Fprintf(result, "%s\n\n", c.SingleLineError(f.Localizer, opts.ColorMode))
	}

	return result.Bytes(), nil
}

func (f SingleLineFormatter) SupportedOutputs() []Output {
	return []Output{OutputChangelog}
}

func getChangelogTitle(changes checker.Changes, l checker.Localizer, colorMode checker.ColorMode) string {
	count := changes.GetLevelCount()
	return l(
		"total-changes",
		len(changes),
		count[checker.ERR],
		checker.ERR.StringCond(colorMode),
		count[checker.WARN],
		checker.WARN.StringCond(colorMode),
		count[checker.INFO],
		checker.INFO.StringCond(colorMode),
	)
}
