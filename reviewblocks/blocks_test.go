package reviewblocks

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestSliceLines(t *testing.T) {
	text := "l1\nl2\nl3\nl4\nl5"
	for _, tc := range []struct {
		name       string
		start, end int
		want       string
	}{
		{"single line", 2, 2, "l2"},
		{"range", 2, 4, "l2\nl3\nl4"},
		{"whole", 1, 5, "l1\nl2\nl3\nl4\nl5"},
		{"clamped end", 4, 99, "l4\nl5"},
		{"out of range start", 99, 100, ""},
		{"inverted", 4, 2, ""},
		{"zero start", 0, 2, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, sliceLines(text, tc.start, tc.end))
		})
	}
}

func TestStructuralKey(t *testing.T) {
	for _, tc := range []struct {
		name           string
		op, path       string
		wantKey, wantT string
	}{
		{"operation", "POST", "/users", "POST /users", "POST /users"},
		{"path level", "", "/users", "/users", "/users"},
		{"component", "", "components/schemas/User", "components/schemas/User", "components/schemas/User"},
		{"fallback", "", "", otherChangesKey, "Other changes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, title := structuralKey(checker.ApiChange{Operation: tc.op, Path: tc.path})
			require.Equal(t, tc.wantKey, key)
			require.Equal(t, tc.wantT, title)
		})
	}
}

// TODO(endpoint-review): once operation/component node resolution lands, add an
// end-to-end test that parses a small spec with IncludeOrigin and asserts the
// sliced block text for each mapping-table case.
