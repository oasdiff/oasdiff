package formatters

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"

	_ "embed"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/report"
)

// GroupEntry is a Group+Changes pair used in sorted template iteration.
type GroupEntry struct {
	Group   ChangeGroup
	Changes *Changes
}

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
		tmpl = template.Must(template.New("changelog").Funcs(HtmlTemplateFuncs()).Parse(changelogHtml))
	}

	return ExecuteHtmlTemplate(tmpl, GroupChanges(changes, f.Localizer), baseVersion, revisionVersion, opts.DiffEmpty)
}

func (f HTMLFormatter) loadCustomTemplate(templatePath string) (*template.Template, error) {
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New("custom-changelog").Funcs(HtmlTemplateFuncs()).Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// changelogTemplateFuncs returns the template functions shared by HTML and Markup changelog templates.
func changelogTemplateFuncs() map[string]any {
	return map[string]any{
		// pathGroups returns path-based groups (ApiChange) sorted by path+operation.
		"pathGroups": func(changes ChangesByGroup) []GroupEntry {
			var entries []GroupEntry
			for g, c := range changes {
				if g.Path != "" {
					entries = append(entries, GroupEntry{g, c})
				}
			}
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Group.Path != entries[j].Group.Path {
					return entries[i].Group.Path < entries[j].Group.Path
				}
				return entries[i].Group.Operation < entries[j].Group.Operation
			})
			return entries
		},
		// sectionGroups returns section-only groups (security, components) sorted by section name.
		"sectionGroups": func(changes ChangesByGroup) []GroupEntry {
			var entries []GroupEntry
			for g, c := range changes {
				if g.Path == "" {
					entries = append(entries, GroupEntry{g, c})
				}
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Group.Section < entries[j].Group.Section
			})
			return entries
		},
		// capitalize returns s with the first letter uppercased.
		"capitalize": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}
}

// HtmlTemplateFuncs returns the FuncMap available to HTML changelog templates.
func HtmlTemplateFuncs() template.FuncMap {
	return template.FuncMap(changelogTemplateFuncs())
}

func ExecuteHtmlTemplate(tmpl *template.Template, changes ChangesByGroup, baseVersion, revisionVersion string, diffEmpty bool) ([]byte, error) {
	var out bytes.Buffer
	if err := tmpl.Execute(&out, TemplateData{changes, baseVersion, revisionVersion, diffEmpty}); err != nil {
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
