package formatters_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gitHubFormatter = formatters.GitHubActionsFormatter{
	Localizer: MockLocalizer,
}

func TestGithubActionsLookup(t *testing.T) {
	f, err := formatters.Lookup(string(formatters.FormatGithubActions), formatters.DefaultFormatterOpts())
	require.NoError(t, err)
	require.IsType(t, formatters.GitHubActionsFormatter{}, f)
}

func TestGitHubActionsFormatter_RenderChangelog_OneFailure(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := "::error title=change_id,file=openapi.yaml::in API GET /api/test This is a breaking change.\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestGitHubActionsFormatter_RenderChangelog_MultipleLevels(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
		checker.ApiChange{
			Id:        "warning_id",
			Level:     checker.WARN,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
		checker.ApiChange{
			Id:        "notice_id",
			Level:     checker.INFO,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := "::error title=change_id,file=openapi.yaml::in API GET /api/test This is a breaking change.\n::warning title=warning_id,file=openapi.yaml::in API GET /api/test This is a warning.\n::notice title=notice_id,file=openapi.yaml::in API GET /api/test This is a notice.\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestGitHubActionsFormatter_DontRenderHttpSource(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("https://google.com/openapi.yaml"),
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := "::error title=change_id::in API GET /api/test This is a breaking change.\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestGitHubActionsFormatter_RenderChangelog_MultilineText(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_two_lines_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := "::error title=change_two_lines_id,file=openapi.yaml::in API GET /api/test This is a breaking change.%0AThis is a second line.\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestGitHubActionsFormatter_RenderChangelog_FileLocation(t *testing.T) {
	testChanges := checker.Changes{
		checker.ApiChange{
			Id:              "change_id",
			Level:           checker.ERR,
			Operation:       http.MethodGet,
			Path:            "/api/test",
			Source:          load.NewSource("openapi.yaml"),
			SourceFile:      "openapi.json",
			SourceLine:      20,
			SourceLineEnd:   25,
			SourceColumn:    5,
			SourceColumnEnd: 10,
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	expectedOutput := "::error title=change_id,file=openapi.json,col=6,endColumn=11,line=21,endLine=26::in API GET /api/test This is a breaking change.\n"
	assert.Equal(t, expectedOutput, string(output))
}

func TestGitHubActionsFormatter_RenderChangelog_JobOutputParameters(t *testing.T) {
	// temp file to mock GITHUB_OUTPUT
	tempFile, err := os.CreateTemp("", "github-output")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())
	_ = os.Setenv("GITHUB_OUTPUT", tempFile.Name())

	testChanges := checker.Changes{
		checker.ApiChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
		checker.ApiChange{
			Id:        "change_id",
			Level:     checker.ERR,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
		checker.ApiChange{
			Id:        "warning_id",
			Level:     checker.WARN,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
		checker.ApiChange{
			Id:        "notice_id",
			Level:     checker.INFO,
			Operation: http.MethodGet,
			Path:      "/api/test",
			Source:    load.NewSource("openapi.yaml"),
		},
	}

	// check output
	output, err := gitHubFormatter.RenderChangelog(testChanges, formatters.NewRenderOpts(), "", "")
	assert.NoError(t, err)
	_ = os.Unsetenv("GITHUB_OUTPUT")
	expectedOutput := "::error title=change_id,file=openapi.yaml::in API GET /api/test This is a breaking change.\n::error title=change_id,file=openapi.yaml::in API GET /api/test This is a breaking change.\n::warning title=warning_id,file=openapi.yaml::in API GET /api/test This is a warning.\n::notice title=notice_id,file=openapi.yaml::in API GET /api/test This is a notice.\n"
	assert.Equal(t, expectedOutput, string(output))

	// check job output parameters (NOTE: order of parameters is not guaranteed)
	outputFile, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(outputFile), "error_count=2\n")
	assert.Contains(t, string(outputFile), "warning_count=1\n")
	assert.Contains(t, string(outputFile), "info_count=1\n")
}

func TestGitHubActionsFormatter_NotImplemented(t *testing.T) {
	var err error
	_, err = gitHubFormatter.RenderDiff(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = gitHubFormatter.RenderSummary(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = gitHubFormatter.RenderChecks(nil, formatters.NewRenderOpts())
	assert.Error(t, err)

	_, err = gitHubFormatter.RenderFlatten(nil, formatters.NewRenderOpts())
	assert.Error(t, err)
}
