package internal

import (
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/build"
)

func TestUploadUserAgent(t *testing.T) {
	t.Setenv("PLATFORM", "")
	t.Setenv("GITHUB_ACTIONS", "")
	if got := uploadUserAgent(); got != "oasdiff-cli/"+build.Version {
		t.Fatalf("plain CLI: %q", got)
	}

	t.Setenv("GITHUB_ACTIONS", "true")
	if got := uploadUserAgent(); !strings.HasSuffix(got, " (github-actions)") {
		t.Fatalf("runner: %q", got)
	}

	// An explicit platform (set by wrappers) wins over runner detection.
	t.Setenv("PLATFORM", "github-action")
	if got := uploadUserAgent(); !strings.HasSuffix(got, " (github-action)") {
		t.Fatalf("explicit platform: %q", got)
	}
}
