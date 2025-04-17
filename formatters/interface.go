package formatters

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/diff"
	"golang.org/x/exp/slices"
)

// Formatter is a common interface for output formatters
type Formatter interface {
	RenderDiff(diff *diff.Diff, opts RenderOpts) ([]byte, error)
	RenderSummary(diff *diff.Diff, opts RenderOpts) ([]byte, error)
	RenderChangelog(changes checker.Changes, opts RenderOpts, baseVersion, revisionVersion string) ([]byte, error)
	RenderChecks(checks Checks, opts RenderOpts) ([]byte, error)
	RenderFlatten(spec *openapi3.T, opts RenderOpts) ([]byte, error)
	SupportedOutputs() []Output
}

var formatters = map[Format]Formatter{
	FormatYAML:          YAMLFormatter{},
	FormatJSON:          JSONFormatter{},
	FormatText:          TEXTFormatter{},
	FormatMarkup:        MarkupFormatter{},
	FormatMarkdown:      MarkupFormatter{},
	FormatSingleLine:    SingleLineFormatter{},
	FormatHTML:          HTMLFormatter{},
	FormatGithubActions: GitHubActionsFormatter{},
	FormatJUnit:         JUnitFormatter{},
}

// Lookup returns a formatter by its name
func Lookup(format string, opts FormatterOpts) (Formatter, error) {
	f := Format(format)
	l := checker.NewLocalizer(opts.Language)

	switch f {
	case FormatYAML:
		return newYAMLFormatter(l), nil
	case FormatJSON:
		return newJSONFormatter(l), nil
	case FormatText:
		return newTEXTFormatter(l), nil
	case FormatMarkup, FormatMarkdown:
		return newMarkupFormatter(l), nil
	case FormatSingleLine:
		return newSingleLineFormatter(l), nil
	case FormatHTML:
		return newHTMLFormatter(l), nil
	case FormatGithubActions:
		return newGitHubActionsFormatter(l), nil
	case FormatJUnit:
		return newJUnitFormatter(l), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", f)
	}
}

func SupportedFormatsByContentType(output Output) []string {
	var formats []string
	for k, v := range formatters {
		if slices.Contains(v.SupportedOutputs(), output) {
			formats = append(formats, string(k))
		}
	}
	sort.Strings(formats)
	return formats
}

// DefaultFormatterOpts returns the default formatter options (e.g. colors, CI mode, etc.)
func DefaultFormatterOpts() FormatterOpts {
	return FormatterOpts{
		Language: localizations.LangDefault,
	}
}
