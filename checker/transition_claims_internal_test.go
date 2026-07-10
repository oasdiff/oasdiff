package checker

import (
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// Every reporter named in the transitions table must be a registered rule: a
// claim whose reporter doesn't exist would silence changes with no finding.
func TestTransitionReportersRegistered(t *testing.T) {
	registered := map[string]bool{}
	for _, rule := range GetAllRules() {
		registered[rule.Id] = true
	}
	for _, tr := range transitions {
		require.NotEmpty(t, tr.reportedBy, "every transition must name its reporters")
		for _, id := range tr.reportedBy {
			require.True(t, registered[id], "transition reporter %q is not a registered rule", id)
		}
	}
}

func TestClaimedByTransition(t *testing.T) {
	nullable := &diff.SchemaDiff{NullableWrappingDiff: &diff.NullableWrappingDiff{NullabilityAdded: true}}
	listOfTypes := &diff.SchemaDiff{ListOfTypesDiff: &diff.ListOfTypesDiff{Added: []string{"integer"}}}
	nullTypeOnly := &diff.SchemaDiff{TypeDiff: &diff.StringsDiff{Added: []string{"null"}}}

	// The nullable wrapping claims every raw reflection of the wrap; kinds are
	// derived from each rule's registry entry.
	require.True(t, claimedByTransition(nullable, RequestPropertyTypeChangedId), "type")
	require.True(t, claimedByTransition(nullable, RequestPropertyEnumValueRemovedId), "values")
	require.True(t, claimedByTransition(nullable, RequestPropertyPatternRemovedId), "constraints")
	require.True(t, claimedByTransition(nullable, RequestPropertyOneOfAddedId), "structure")
	require.True(t, claimedByTransition(nullable, RequestPropertyRemovedId),
		"existence: an object's properties read as removed when it is wrapped")
	require.True(t, claimedByTransition(nullable, RequestPropertyBecameRequiredId),
		"requiredness: required entries move between the top level and the branch")
	require.False(t, claimedByTransition(nullable, RequestPropertyDeprecatedId),
		"lifecycle changes are not reflections of the wrap and keep reporting")

	// A list-of-types transition claims the type and structure reflections but
	// not values: an enum change there is a real change.
	require.True(t, claimedByTransition(listOfTypes, RequestPropertyTypeChangedId))
	require.True(t, claimedByTransition(listOfTypes, RequestPropertyOneOfAddedId))
	require.False(t, claimedByTransition(listOfTypes, RequestPropertyEnumValueRemovedId))

	// A null-only type change claims only the type reflection.
	require.True(t, claimedByTransition(nullTypeOnly, RequestPropertyTypeChangedId))
	require.False(t, claimedByTransition(nullTypeOnly, RequestPropertyOneOfAddedId))

	// The reportedBy exemption: a transition never claims its own reporters,
	// regardless of the kind they are registered under.
	require.False(t, claimedByTransition(listOfTypes, RequestPropertyListOfTypesWidenedId),
		"a KindType reporter at its own KindType-claiming transition still reports")
	require.True(t, claimedByTransition(nullable, RequestParameterListOfTypesWidenedId),
		"the parameter list-of-types finding yields to the dedicated parameter became-nullable reporter")
	require.False(t, claimedByTransition(nullable, RequestParameterBecameNullableId),
		"the parameter became-nullable finding is the nullable transition's parameter-level reporter")
	require.True(t, claimedByTransition(nullable, RequestPropertyListOfTypesWidenedId),
		"the property list-of-types finding yields to became-nullable")

	// Missing plumbing never claims: over-reporting, not a lost finding.
	require.False(t, claimedByTransition(nil, RequestPropertyEnumValueRemovedId), "no schema node")
	require.False(t, claimedByTransition(nullable, "no-such-rule"), "unregistered id")
	require.False(t, claimedByTransition(&diff.SchemaDiff{}, RequestPropertyTypeChangedId), "no transition present")

	// WithSchema is the construction-time entry point for the decision.
	require.True(t, ApiChange{Id: RequestPropertyEnumValueRemovedId}.WithSchema(nullable).claimed)
	require.False(t, ApiChange{Id: RequestPropertyEnumValueRemovedId}.WithSchema(nil).claimed)
}
