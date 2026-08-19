package checker_test

import (
	"os"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
)

const coverageDocPath = "../docs/COVERAGE.md"

// TestCoverageDoc verifies that docs/COVERAGE.md matches the rule registry.
// Regenerate with: make coverage-doc
func TestCoverageDoc(t *testing.T) {
	want := checker.CoverageDoc()
	if os.Getenv("UPDATE_COVERAGE_DOC") != "" {
		if err := os.WriteFile(coverageDocPath, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile(coverageDocPath)
	if err != nil {
		t.Fatalf("%v; generate the file with: make coverage-doc", err)
	}
	// a Windows checkout may convert the file to CRLF
	if strings.ReplaceAll(string(got), "\r\n", "\n") != want {
		t.Error("docs/COVERAGE.md is out of date with the rule registry; regenerate with: make coverage-doc")
	}
}
