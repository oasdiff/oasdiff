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

func (f GitHubActionsFormatter) RenderChangelog(changes checker.Changes, opts RenderOpts, _, _ string) ([]byte, error) {
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
			"title=" + change.GetId(),
		}
		if rev := change.GetRevisionSource(); rev != nil && rev.File != "" && !isHTTPSource(rev.File) {
			params = append(params, "file="+rev.File)
			if rev.Line != 0 {
				params = append(params, "line="+strconv.Itoa(rev.Line))
			}
			if rev.Column != 0 {
				params = append(params, "col="+strconv.Itoa(rev.Column))
			}
		} else if sourceFile := change.GetSourceFile(); sourceFile != "" {
			// Fallback: operation source file when no revision source is available
			params = append(params, "file="+sourceFile)
		}

		buf.WriteString(fmt.Sprintf("::%s %s::%s\n", githubActionsSeverity[change.GetLevel()], strings.Join(params, ","), getMessage(change, f.Localizer)))
	}

	return buf.Bytes(), nil
}

func getMessage(change checker.Change, l checker.Localizer) string {
	message := strings.ReplaceAll(change.GetUncolorizedText(l), "\n", "%0A")
	return fmt.Sprintf("in API %s %s %s", change.GetOperation(), change.GetPath(), message)
}

func (f GitHubActionsFormatter) SupportedOutputs() []Output {
	return []Output{OutputChangelog}
}

func isHTTPSource(file string) bool {
	return strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://")
}

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
	defer file.Close()

	// collect all parameters into a string
	var contentBuilder strings.Builder
	for key, value := range params {
		contentBuilder.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	// write the parameters to the file
	if _, err := file.WriteString(contentBuilder.String()); err != nil {
		return fmt.Errorf("failed to write GitHub Actions job output parameters: %w", err)
	}

	return nil
}
