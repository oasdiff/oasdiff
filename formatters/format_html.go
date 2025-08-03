package formatters

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

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

func (f HTMLFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, baseVersion, revisionVersion string) ([]byte, error) {
	var tmpl *template.Template
	var err error

	if opts.TemplatePath != "" {
		tmpl, err = f.loadCustomTemplate(opts.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom template: %w", err)
		}
	} else {
		tmpl = template.Must(template.New("changelog").Parse(changelogHtml))
	}

	return ExecuteHtmlTemplate(tmpl, GroupChanges(changes, f.Localizer), baseVersion, revisionVersion)
}

func (f HTMLFormatter) loadCustomTemplate(templatePath string) (*template.Template, error) {
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("custom-changelog").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
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

func (f HTMLFormatter) SupportsTemplate() bool {
	return true
}
