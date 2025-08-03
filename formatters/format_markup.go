package formatters

import (
	"bytes"
	"fmt"
	"os"
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
	var tmpl *template.Template
	var err error

	if opts.TemplatePath != "" {
		tmpl, err = f.loadCustomTemplate(opts.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom template: %w", err)
		}
	} else {
		tmpl = template.Must(template.New("changelog").Parse(changelogMarkdown))
	}

	return ExecuteTextTemplate(tmpl, GroupChanges(changes, f.Localizer), baseVersion, revisionVersion)
}

func (f MarkupFormatter) loadCustomTemplate(templatePath string) (*template.Template, error) {
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

func (f MarkupFormatter) SupportsTemplate() bool {
	return true
}
