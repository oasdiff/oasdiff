package checker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The all-of-not-flattened disclaimer says a sibling branch may still
// guarantee what this one dropped, so it belongs only on a removed branch
// that constrained something. Driven off the rule set, so a new *-all-of-*
// id is covered here the day it is added.
func TestAllOfDisclaimerOnRemovedBranchesOnly(t *testing.T) {
	var ids []string
	for _, rule := range GetAllRules() {
		if strings.Contains(rule.Id, "-all-of-") {
			ids = append(ids, rule.Id)
		}
	}
	require.Len(t, ids, 16)

	for _, id := range ids {
		got := allOfDisclaimers(true, nil, id)
		if strings.HasSuffix(id, "-all-of-removed") {
			require.Equal(t, []Disclaimer{DisclaimerAllOfNotFlattened}, got, id)
		} else {
			require.Empty(t, got, id)
		}
	}
}
