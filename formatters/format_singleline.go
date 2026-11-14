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

func (f SingleLineFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts) ([]byte, error) {
	result := bytes.NewBuffer(nil)

	if len(changes) == 0 {
		if opts.DiffEmpty {
			_, _ = fmt.Fprint(result, "No changes detected")
		} else if opts.IsBreaking {
			_, _ = fmt.Fprint(result, "No breaking changes to report, but the specs are different. Run 'oasdiff diff' to see structural differences.")
		} else {
			_, _ = fmt.Fprint(result, "No changes to report, but the specs are different. Run 'oasdiff diff' to see structural differences.")
		}
		return result.Bytes(), nil
	}

	_, _ = fmt.Fprint(result, getChangelogTitle(changes, f.Localizer, opts.ColorMode))

	for _, c := range changes {
		_, _ = fmt.Fprintf(result, "%s\n\n", singleLineError(c, f.Localizer, opts.ColorMode))
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
		levelText(checker.ERR, colorMode),
		count[checker.WARN],
		levelText(checker.WARN, colorMode),
		count[checker.INFO],
		levelText(checker.INFO, colorMode),
	)
}
