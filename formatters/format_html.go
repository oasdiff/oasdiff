package formatters

import (
	"bytes"
	"fmt"
	"html/template"

	_ "embed"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

type HTMLFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newHTMLFormatter(l checker.Localizer) HTMLFormatter {
	return HTMLFormatter{
		Localizer: l,
	}
}

func (f HTMLFormatter) RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error) {
	reportAsString, err := report.GetHTMLReportAsString(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML report: %w", err)
	}

	return []byte(reportAsString), nil
}

//go:embed templates/changelog.html
var changelogHtml string

type TemplateData struct {
	APIChanges      ChangesByEndpoint
	BaseVersion     string
	RevisionVersion string
}

func (t TemplateData) GetVersionTitle() string {
	if t.BaseVersion == "" || t.RevisionVersion == "" {
		return ""
	}

	return fmt.Sprintf("%s vs. %s", t.BaseVersion, t.RevisionVersion)
}

func (f HTMLFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, baseVersion, revisionVersion string) ([]byte, error) {
	tmpl := template.Must(template.New("changelog").Parse(changelogHtml))
	return ExecuteHtmlTemplate(tmpl, GroupChanges(changes, f.Localizer), baseVersion, revisionVersion)
}

func ExecuteHtmlTemplate(tmpl *template.Template, changes ChangesByEndpoint, baseVersion, revisionVersion string) ([]byte, error) {
	var out bytes.Buffer
	if err := tmpl.Execute(&out, TemplateData{changes, baseVersion, revisionVersion}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (f HTMLFormatter) SupportedOutputs() []Output {
	return []Output{OutputDiff, OutputChangelog}
}
