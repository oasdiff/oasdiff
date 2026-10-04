package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// sharedSchemaChanges runs one check over a fixture pair under
// data/checker, named without its _base or _revision suffix.
func sharedSchemaChanges(t *testing.T, fixture string, check checker.BackwardCompatibilityCheck) checker.Changes {
	t.Helper()
	base, err := open("../data/checker/" + fixture + "_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/" + fixture + "_revision.yaml")
	require.NoError(t, err)
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	return checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(check), d, sources, checker.INFO)
}

// sharedAt returns, for each change, the property it is reported at and the
// shared schema attached to it.
func sharedAt(t *testing.T, changes checker.Changes) map[string]*checker.SharedSchema {
	t.Helper()
	result := map[string]*checker.SharedSchema{}
	for _, change := range changes {
		apiChange := change.(checker.ApiChange)
		require.NotNil(t, apiChange.GetSharedSchema(), "%s at %v has no shared schema", apiChange.Id, apiChange.Args)
		result[apiChange.GetSharedSchema().Properties[0]] = apiChange.GetSharedSchema()
	}
	return result
}

// The walk reaches Shared through left only, and the change is reported at
// right too, each naming the other.
func TestSharedSchemaChangeIsReportedAtEachReference(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/extra":  {Name: "Shared", Properties: []string{"left/extra", "right/extra"}},
		"right/extra": {Name: "Shared", Properties: []string{"right/extra", "left/extra"}},
	}, sharedAt(t, changes))
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Shared, also at `right/extra`)")
}

// The copies come out in the same order on every run.
func TestSharedSchemaChangeOrderIsStable(t *testing.T) {
	first := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)
	for range 30 {
		require.Equal(t, first, sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck))
	}
}

// userId and customerId mean different things but share Id, so a pattern
// added to Id is reported at both: one property of a schema referenced from
// the same parent twice is not left out.
func TestSharedSchemaChangeIsReportedAtEachPropertyOfOneParent(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_two_properties", checker.ResponsePatternAddedOrChangedCheck)

	operations := map[string][]string{}
	for _, change := range changes {
		shared := change.(checker.ApiChange).GetSharedSchema()
		require.NotNil(t, shared)
		operations[change.GetPath()] = append(operations[change.GetPath()], shared.Properties[0])
	}
	require.Equal(t, map[string][]string{
		"/orders":   {"customerId", "userId"},
		"/accounts": {"customerId", "ownerId", "userId"},
	}, operations)
}

// Shared is reached through First and Second, which are different parents,
// and each of those through two properties of the payload. The list has one
// entry per reference to Shared, so the paths through the second use of
// First and of Second are not repeated.
func TestSharedSchemaUnderTwoParentsIsReportedAtEachReference(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_two_parents", checker.ResponseOptionalPropertyUpdatedCheck)

	inShared := checker.Changes{}
	for _, change := range changes {
		if change.(checker.ApiChange).GetSharedSchema() != nil && change.(checker.ApiChange).GetSharedSchema().Name == "Shared" {
			inShared = append(inShared, change)
		}
	}
	require.Equal(t, map[string]*checker.SharedSchema{
		"first/shared/extra":  {Name: "Shared", Properties: []string{"first/shared/extra", "second/shared/extra"}},
		"second/shared/extra": {Name: "Shared", Properties: []string{"second/shared/extra", "first/shared/extra"}},
	}, sharedAt(t, inShared))
}

// A change below the shared schema, not in it, is reported at each reference
// too, with the rest of its path appended.
func TestChangeBelowASharedSchemaIsReportedAtEachReference(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_inner_change", checker.ResponsePropertyTypeChangedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/id":  {Name: "Shared", Properties: []string{"left/id", "right/id"}},
		"right/id": {Name: "Shared", Properties: []string{"right/id", "left/id"}},
	}, sharedAt(t, changes))
}

// A schema reached through a JSON pointer into another schema has no
// components.schemas name, and the walk passes through none on the way to it.
func TestSharedSchemaWithoutANameIsReportedAtEachReference(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_unnamed", checker.ResponseOptionalPropertyUpdatedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/extra":  {Properties: []string{"left/extra", "right/extra"}},
		"right/extra": {Properties: []string{"right/extra", "left/extra"}},
	}, sharedAt(t, changes))
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(also at `right/extra`)")
}

// A description beside each $ref to Address makes the parser copy Address
// once per property. The copies share their children, so zip is what both
// properties reach, and it is named after Address, the schema it belongs to.
func TestSharedSchemaWrittenInlineIsNamedAfterItsComponent(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_override", checker.ResponsePatternAddedOrChangedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"billing/zip":  {Name: "Address", Properties: []string{"billing/zip", "shipping/zip"}},
		"shipping/zip": {Name: "Address", Properties: []string{"shipping/zip", "billing/zip"}},
	}, sharedAt(t, changes))
}

// A schema used once gets no shared schema.
func TestSchemaUsedOnceHasNoSharedSchema(t *testing.T) {
	base, err := open("../data/component-renamed1.yaml")
	require.NoError(t, err)
	revision, err := open("../data/component-renamed2.yaml")
	require.NoError(t, err)
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck), d, sources, checker.INFO)

	require.Len(t, changes, 1)
	require.Nil(t, changes[0].(checker.ApiChange).GetSharedSchema())
}
