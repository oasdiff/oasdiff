package formatters_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// fingerprintOf is a small helper that runs a single change through NewChanges
// and returns the resulting fingerprint. The MockLocalizer keeps the rendered
// text constant so any drift the test catches is in the fingerprint logic
// itself, not the message catalogue.
func fingerprintOf(c checker.Change) string {
	return formatters.NewChanges(checker.Changes{c}, MockLocalizer)[0].Fingerprint
}

// Pin the current fingerprint format so a future change to the algorithm
// (e.g. switching hash, truncation length, or the order of inputs) becomes a
// loud test failure rather than a silent invalidation of every stored
// pr_changes row in oasdiff-service.
func TestNewChanges_FingerprintFormat(t *testing.T) {
	got := fingerprintOf(checker.ApiChange{
		Id:        "change_id",
		Level:     checker.ERR,
		Operation: "GET",
		Path:      "/users",
		Source:    &load.Source{},
		Args:      []any{"name", 42},
	})
	require.Equal(t, "39d9aa2ed55d", got)
	require.Len(t, got, 12)
}

// Args go through fmt's %v verb, so different values must hash differently.
// This is the disambiguation property: two changes from the same rule on the
// same operation with different parameters must not collide.
func TestNewChanges_FingerprintDistinguishesArgs(t *testing.T) {
	a := fingerprintOf(checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "GET", Path: "/u",
		Source: &load.Source{}, Args: []any{"alpha"},
	})
	b := fingerprintOf(checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "GET", Path: "/u",
		Source: &load.Source{}, Args: []any{"beta"},
	})
	require.NotEqual(t, a, b)
}

// The whole point of the change: the fingerprint must not depend on the
// rendered text. This is the regression guard for the April 9 backticks
// incident — a copy edit to the localized message would silently invalidate
// every stored fingerprint under the old text-based algorithm.
//
// We compare two changes that produce the same args+id+op+path but render
// different text (different rule families with overlapping localization keys
// would differ on text alone). Using the same change but a different
// Localizer gives us a tighter, locale-only test.
func TestNewChanges_FingerprintIndependentOfLocale(t *testing.T) {
	change := checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "POST", Path: "/x",
		Source: &load.Source{}, Args: []any{"v1"},
	}
	english := formatters.NewChanges(checker.Changes{change}, MockLocalizer)[0]
	other := formatters.NewChanges(checker.Changes{change}, func(originalKey string, args ...interface{}) string {
		return "completely different rendering: " + originalKey
	})[0]

	require.NotEqual(t, english.Text, other.Text, "test setup: locales must render differently")
	require.Equal(t, english.Fingerprint, other.Fingerprint,
		"fingerprint must not depend on rendered text — copy edits and locale switches would invalidate every stored pr_changes row in oasdiff-service")
}

// Two changes that differ only in BaseSource line/column must share a
// fingerprint so review state survives unrelated whitespace edits to the spec
// (the carry-forward use case in oasdiff-service/internal/pr_comment.go).
func TestNewChanges_FingerprintIndependentOfSourceLocation(t *testing.T) {
	a := fingerprintOf(checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "GET", Path: "/u",
		Source: &load.Source{}, Args: []any{"p"},
		CommonChange: checker.CommonChange{
			BaseSource: checker.NewSource("base.yaml", 10, 5),
		},
	})
	b := fingerprintOf(checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "GET", Path: "/u",
		Source: &load.Source{}, Args: []any{"p"},
		CommonChange: checker.CommonChange{
			BaseSource: checker.NewSource("base.yaml", 99, 5),
		},
	})
	require.Equal(t, a, b)
}

// A change with no args gets an empty arg segment, not a panic.
func TestNewChanges_FingerprintNoArgs(t *testing.T) {
	got := fingerprintOf(checker.ApiChange{
		Id: "change_id", Level: checker.ERR, Operation: "GET", Path: "/u",
		Source: &load.Source{},
	})
	require.Len(t, got, 12)
}
