package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

// One edit can be a transition and an unrelated real change at the same node:
// here the type widens into a oneOf (the list-of-types transition) AND the
// enum constraint is dropped. The transition claims only the type and
// structure echoes, so the enum findings must survive alongside the
// transition's own finding. This is what scoping suppression by claimed kind
// buys: if presence alone suppressed (with only the reportedBy exemption),
// the enum findings would be silently lost.
func TestTransitionClaims_RealChangeAlongsideTransitionStillReported(t *testing.T) {
	ids := nullableWrapChanges(t, "../data/checker/transition_claims_enum_base.yaml", "../data/checker/transition_claims_enum_revision.yaml")
	require.ElementsMatch(t, []string{
		checker.RequestPropertyListOfTypesWidenedId, // the transition's finding
		checker.RequestPropertyEnumValueRemovedId,   // alpha: a real change, not an echo
		checker.RequestPropertyEnumValueRemovedId,   // beta
	}, ids)
}
