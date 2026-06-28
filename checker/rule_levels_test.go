package checker_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// goldenRuleLevels pins each rule id's default level (the id->level contract that
// per-test Level assertions would otherwise duplicate).
// Regenerate: UPDATE_GOLDEN=1 go test ./checker -run TestRuleLevels
const goldenRuleLevels = "../data/rule_levels.txt"

func TestRuleLevels(t *testing.T) {
	got := formatRuleLevels(allChecksConfig().LogLevels)

	if os.Getenv("UPDATE_GOLDEN") != "" {
		require.NoError(t, os.WriteFile(goldenRuleLevels, []byte(got), 0o644))
	}

	want, err := os.ReadFile(goldenRuleLevels)
	require.NoError(t, err)
	require.Equal(t, string(want), got,
		"rule levels changed; if intentional, regenerate with UPDATE_GOLDEN=1 go test ./checker -run TestRuleLevels")
}

func formatRuleLevels(levels map[string]checker.Level) string {
	ids := make([]string, 0, len(levels))
	for id := range levels {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	for _, id := range ids {
		fmt.Fprintf(&b, "%s %s\n", id, levels[id].String())
	}
	return b.String()
}
