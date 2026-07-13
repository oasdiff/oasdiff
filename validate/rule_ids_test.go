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

// A derived ID outside the registry is demoted to spec-validation-error: an
// upstream rename or a new validation surfaces loudly instead of silently
// minting a new public ID.
func TestKnownRuleID_GatesUnregisteredIDs(t *testing.T) {
	renamed := ruleIDForKinError(&openapi3.RequiredFieldError{Field: "someRenamedKinField"})
	require.Equal(t, "some-renamed-kin-field-required", renamed, "derivation still runs")
	require.Equal(t, unknownValidationID, knownRuleID(renamed), "unregistered id is demoted")

	known := ruleIDForKinError(&openapi3.RequiredFieldError{Field: "oAuthFlow.scopes"})
	require.Equal(t, "oauth-flow-scopes-required", knownRuleID(known), "registered id passes through")
}

// A 3.2-gated field reports a 3-2-plus id, not 3-1-plus (kin's MinVersion
// drives the suffix).
func TestRuleID_VersionMismatchSuffixFollowsMinVersion(t *testing.T) {
	id := ruleIDForKinError(&openapi3.FieldVersionMismatchError{Field: "itemSchema", MinVersion: "3.2"})
	require.Equal(t, "item-schema-field-for-3-2-plus", id)
	require.Equal(t, id, knownRuleID(id), "the 3.2 id is registered")
}
