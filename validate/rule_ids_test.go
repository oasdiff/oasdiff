package validate

import (
	"regexp"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// The registry is the public ID surface: sorted (knownRuleID binary-searches
// it), duplicate-free, kebab-case.
func TestRuleIDs_Registry(t *testing.T) {
	require.True(t, slices.IsSorted(ruleIDs), "registry must be sorted")
	require.Equal(t, len(slices.Compact(slices.Clone(ruleIDs))), len(ruleIDs), "registry must have no duplicates")
	kebab := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, id := range ruleIDs {
		require.Truef(t, kebab.MatchString(id), "rule id %q is not kebab-case", id)
	}
	require.Contains(t, ruleIDs, unknownValidationID)
	require.Contains(t, ruleIDs, DuplicateEnumValueID)
}

// The registry is exactly kin's declared validation-error codes plus the
// oasdiff-native ids. A kin bump that adds or renames a code fails this until
// the registry is updated deliberately, so the public ID surface never drifts
// silently. This replaces the old per-family derivation checks: the ids now
// come from kin's CodedError.Code(), so kin owns the spellings and this pins
// only that oasdiff's registry agrees with them.
func TestRuleIDs_MatchKinCatalog(t *testing.T) {
	want := slices.Concat(openapi3.ValidationErrorCodes(), nativeRuleIDs)
	slices.Sort(want)
	require.Equal(t, want, ruleIDs,
		"registry and kin catalog diverge: reconcile ruleIDs with openapi3.ValidationErrorCodes() ∪ nativeRuleIDs")
}

// A code the registry does not know is demoted to spec-validation-error; a
// known code passes through. This guards oasdiff's gate; the codes themselves
// are kin's (see TestRuleIDs_MatchKinCatalog).
func TestKnownRuleID_GatesUnregisteredCodes(t *testing.T) {
	const unregistered = "some-code-kin-does-not-emit"
	require.NotContains(t, ruleIDs, unregistered)
	require.Equal(t, unknownValidationID, knownRuleID(unregistered), "an unregistered code is demoted")
	require.Equal(t, "oauth-flow-scopes-required", knownRuleID("oauth-flow-scopes-required"), "a registered code passes through")
}
