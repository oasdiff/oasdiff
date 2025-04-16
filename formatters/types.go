package formatters

import "github.com/oasdiff/oasdiff/checker"

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
	WrapInObject bool // wrap the output in a JSON object with the key "changes"
}

func NewRenderOpts() RenderOpts {
	return RenderOpts{
		ColorMode: checker.ColorAuto,
	}
}
