package formatters

import (
	"bytes"
	"text/template"

	_ "embed"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

type MarkupFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newMarkupFormatter(l checker.Localizer) MarkupFormatter {
	return MarkupFormatter{
		Localizer: l,
	}
}

func (f MarkupFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	return []byte(report.GetTextReportAsString(diff)), nil
}

//go:embed templates/changelog.md
var changelogMarkdown string

func (f MarkupFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, baseVersion, revisionVersion string) ([]byte, error) {
	tmpl := template.Must(template.New("changelog").Parse(changelogMarkdown))
	return ExecuteTextTemplate(tmpl, GroupChanges(changes, f.Localizer), baseVersion, revisionVersion)
}

func ExecuteTextTemplate(tmpl *template.Template, changes ChangesByEndpoint, baseVersion, revisionVersion string) ([]byte, error) {
	var out bytes.Buffer
	if err := tmpl.Execute(&out, TemplateData{changes, baseVersion, revisionVersion}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (f MarkupFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputChangelog}
}
