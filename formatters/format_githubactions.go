package formatters

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
)

var githubActionsSeverity = map[checker.Level]string{
	checker.ERR:  "error",
	checker.WARN: "warning",
	checker.INFO: "notice",
}

type GitHubActionsFormatter struct {
	notImplementedFormatter
	Localizer checker.Localizer
}

func newGitHubActionsFormatter(l checker.Localizer) GitHubActionsFormatter {
	return GitHubActionsFormatter{
		Localizer: l,
	}
}

func (f GitHubActionsFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts) ([]byte, error) {
	var buf bytes.Buffer

	// add error, warning and notice count to job output parameters
	err := writeGitHubActionsJobOutputParameters(map[string]string{
		"error_count":   fmt.Sprint(changes.GetLevelCount()[checker.ERR]),
		"warning_count": fmt.Sprint(changes.GetLevelCount()[checker.WARN]),
		"info_count":    fmt.Sprint(changes.GetLevelCount()[checker.INFO]),
	})
	if err != nil {
		return nil, err
	}

	// generate messages for each change (source file, line and column are optional)
	for _, change := range changes {
		var params = []string{
			"title=" + escapeProperty(change.GetId()),
		}
		if rev := change.GetRevisionSource(); rev != nil && rev.File != "" && !isHTTPSource(rev.File) {
			params = append(params, "file="+escapeProperty(rev.File))
			if rev.Line != 0 {
				params = append(params, "line="+strconv.Itoa(rev.Line))
			}
			if rev.Column != 0 {
				params = append(params, "col="+strconv.Itoa(rev.Column))
			}
		} else if sourceFile := change.GetSourceFile(); sourceFile != "" {
			// Fallback: operation source file when no revision source is available
			params = append(params, "file="+escapeProperty(sourceFile))
		}

		fmt.Fprintf(&buf, "::%s %s::%s\n", githubActionsSeverity[change.GetLevel()], strings.Join(params, ","), getMessage(change, f.Localizer))
	}

	return buf.Bytes(), nil
}

func getMessage(change checker.Change, l checker.Localizer) string {
	return escapeData(fmt.Sprintf("in API %s %s %s", change.GetOperation(), change.GetPath(), change.GetUncolorizedText(l)))
}

// escapeData escapes a string for use as a GitHub Actions annotation
// message, per the workflow-command rules. % is escaped first so the
// %0D/%0A sequences it introduces are not themselves double-escaped.
func escapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}

// escapeProperty escapes a string for use as a GitHub Actions annotation
// property value (title, file): escapeData plus the property delimiters
// ':' and ',' (e.g. a git-ref source like "main:openapi.yaml").
func escapeProperty(s string) string {
	s = escapeData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}

// RenderValidate emits one GitHub Actions annotation per finding so the
// oasdiff-action validate wrapper can surface violations inline on the
// PR's Files Changed tab, and publishes per-severity counts as step
// outputs. Findings carry exact file/line/column, which the annotation
// uses to anchor itself.
func (f GitHubActionsFormatter) RenderValidate(findings Findings, opts RenderOpts) ([]byte, error) {
	var buf bytes.Buffer

	count := findings.GetLevelCount()
	err := writeGitHubActionsJobOutputParameters(map[string]string{
		"error_count":   fmt.Sprint(count[checker.ERR]),
		"warning_count": fmt.Sprint(count[checker.WARN]),
		"info_count":    fmt.Sprint(count[checker.INFO]),
	})
	if err != nil {
		return nil, err
	}

	for _, finding := range findings {
		params := []string{"title=" + escapeProperty(finding.Id)}
		if finding.Source.File != "" && !isHTTPSource(finding.Source.File) {
			params = append(params, "file="+escapeProperty(finding.Source.File))
			if finding.Source.Line != 0 {
				params = append(params, "line="+strconv.Itoa(finding.Source.Line))
			}
			if finding.Source.Column != 0 {
				params = append(params, "col="+strconv.Itoa(finding.Source.Column))
			}
		}

		fmt.Fprintf(&buf, "::%s %s::%s\n", githubActionsSeverity[finding.Level], strings.Join(params, ","), getFindingMessage(finding))
	}

	return buf.Bytes(), nil
}

// getFindingMessage renders a finding's annotation body. The "in API
// METHOD /path" prefix is added only when the finding has operation
// context; doc-root and components-rooted findings (the majority) have
// none, so prefixing would emit a stray "in API  ".
func getFindingMessage(finding Finding) string {
	message := finding.Text
	if finding.Operation != "" && finding.Path != "" {
		message = fmt.Sprintf("in API %s %s %s", finding.Operation, finding.Path, message)
	}
	return escapeData(message)
}

func (f GitHubActionsFormatter) SupportedOutputs() []Output {
	return []Output{OutputChangelog, OutputValidate}
}

func isHTTPSource(file string) bool {
	return strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://")
}

// writeGitHubActionsJobOutputParameters publishes key=value pairs as GitHub Actions
// step output parameters by appending them to the file referenced by $GITHUB_OUTPUT.
// This is the mechanism GitHub Actions uses to pass outputs between steps
// (https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/passing-information-between-jobs).
// If $GITHUB_OUTPUT is unset (i.e. running outside GitHub Actions), this is a no-op.
func writeGitHubActionsJobOutputParameters(params map[string]string) error {
	githubOutputFile := os.Getenv("GITHUB_OUTPUT")
	if githubOutputFile == "" {
		// If GITHUB_OUTPUT is not set, we can't write job output parameters (running outside of GitHub Actions)
		return nil
	}

	// open the file in append mode
	file, err := os.OpenFile(githubOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GitHub Actions job output file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// collect all parameters into a string
	var contentBuilder strings.Builder
	for key, value := range params {
		fmt.Fprintf(&contentBuilder, "%s=%s\n", key, value)
	}

	// write the parameters to the file
	if _, err := file.WriteString(contentBuilder.String()); err != nil {
		return fmt.Errorf("failed to write GitHub Actions job output parameters: %w", err)
	}

	return nil
}
