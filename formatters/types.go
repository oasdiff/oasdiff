package formatters

import (
	"fmt"
	"github.com/oasdiff/oasdiff/checker"
)

type Format string

const (
	FormatYAML          Format = "yaml"
	FormatJSON          Format = "json"
	FormatText          Format = "text"
	FormatMarkup        Format = "markup"
	FormatMarkdown      Format = "markdown"
	FormatSingleLine    Format = "singleline"
	FormatHTML          Format = "html"
	FormatGithubActions Format = "githubactions"
	FormatJUnit         Format = "junit"
	FormatSarif         Format = "sarif"
)

func GetSupportedFormats() []string {
	return []string{
		string(FormatYAML),
		string(FormatJSON),
		string(FormatText),
		string(FormatMarkup),
		string(FormatMarkdown),
		string(FormatSingleLine),
		string(FormatHTML),
		string(FormatGithubActions),
		string(FormatJUnit),
		string(FormatSarif),
	}
}

// FormatterOpts can be used to pass properties to the formatter (e.g. colors)
type FormatterOpts struct {
	Language string
}

// RenderOpts can be used to pass properties to the renderer method
type RenderOpts struct {
	ColorMode    checker.ColorMode
	WrapInObject bool   // wrap the output in a JSON object with the key "changes"
	TemplatePath string // path to custom template file for changelog generation
	DiffEmpty    bool   // true when the underlying diff found no changes at all
	IsBreaking   bool   // true when invoked via `oasdiff breaking` (vs `changelog`); affects empty-result wording
}

func NewRenderOpts() RenderOpts {
	return RenderOpts{
		ColorMode: checker.ColorAuto,
	}
}

type TemplateData struct {
	GroupedChanges  ChangesByGroup
	BaseVersion     string
	RevisionVersion string
	DiffEmpty       bool
	IsBreaking      bool
}

// APIChanges returns GroupedChanges.
// Deprecated: Use .GroupedChanges in templates instead. Kept for backward compatibility with custom templates.
func (t TemplateData) APIChanges() ChangesByGroup {
	return t.GroupedChanges
}

func (t TemplateData) GetVersionTitle() string {
	if t.BaseVersion == "" || t.RevisionVersion == "" {
		return ""
	}

	return fmt.Sprintf("%s vs. %s", t.BaseVersion, t.RevisionVersion)
}
