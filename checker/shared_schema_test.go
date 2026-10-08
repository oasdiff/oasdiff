package checker_test

import (
	"testing"

	"github.com/TwiN/go-color"
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
		require.NotContains(t, result, apiChange.GetSharedSchema().Properties[0])
		result[apiChange.GetSharedSchema().Properties[0]] = apiChange.GetSharedSchema()
	}
	return result
}

// The walk reaches Shared through left only, so the change is reported at
// left, naming right as well.
func TestSharedSchemaChangeListsTheOtherProperty(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/extra": {Name: "Shared", Properties: []string{"left/extra", "right/extra"}, Count: 2},
	}, sharedAt(t, changes))
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: `Shared`, also at `right/extra`)")
}

// In colored output, the values in the shared schema detail are formatted the
// same way as the message arguments.
func TestSharedSchemaDetailIsColoredLikeTheArguments(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)

	text := changes[0].GetText(checker.NewLocalizer("en"))
	require.Contains(t, text, color.InBold("'left/extra'"))
	require.Contains(t, text, "(shared schema: "+color.InBold("'Shared'")+", also at "+color.InBold("'right/extra'")+")")
}

func TestSharedSchemaDetailIsLocalized(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)

	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("es")), "(esquema compartido: `Shared`, también en `right/extra`)")
}

// The properties come out in the same order on every run.
func TestSharedSchemaChangeOrderIsStable(t *testing.T) {
	first := sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck)
	for range 30 {
		require.Equal(t, first, sharedSchemaChanges(t, "shared_schema", checker.ResponseOptionalPropertyUpdatedCheck))
	}
}

// userId and customerId mean different things but share Id, so a pattern
// added to Id names both: one property of a schema referenced from the same
// parent twice is not left out. Each operation reports it once.
func TestSharedSchemaChangeListsEachPropertyOfOneParent(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_two_properties", checker.ResponsePatternAddedOrChangedCheck)

	operations := map[string]*checker.SharedSchema{}
	for _, change := range changes {
		require.NotContains(t, operations, change.GetPath())
		operations[change.GetPath()] = change.(checker.ApiChange).GetSharedSchema()
	}
	require.Equal(t, map[string]*checker.SharedSchema{
		"/orders":   {Name: "Id", Properties: []string{"customerId", "userId"}, Count: 2},
		"/accounts": {Name: "Id", Properties: []string{"customerId", "ownerId", "userId"}, Count: 3},
	}, operations)
}

// Shared is reached through First and Second, which are different parents,
// and each of those through two properties of the payload, so the change is at
// four properties. The first three are listed, and the count includes the
// fourth.
func TestSharedSchemaUnderTwoParentsCountsEveryProperty(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_two_parents", checker.ResponseOptionalPropertyUpdatedCheck)

	inShared := checker.Changes{}
	for _, change := range changes {
		if change.(checker.ApiChange).GetSharedSchema() != nil && change.(checker.ApiChange).GetSharedSchema().Name == "Shared" {
			inShared = append(inShared, change)
		}
	}
	require.Equal(t, map[string]*checker.SharedSchema{
		"first/shared/extra": {Name: "Shared", Properties: []string{"first/shared/extra", "third/shared/extra", "fourth/shared/extra"}, Count: 4},
	}, sharedAt(t, inShared))
	require.Contains(t, inShared[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: `Shared`, also at `third/shared/extra`, `fourth/shared/extra` and 1 more)")
}

// Inner is shared inside Outer, which is itself shared, so the change is at
// every combination of the two: p/a, q/a, p/b and q/b.
func TestNestedSharedSchemaCountsEveryProperty(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_nested", checker.ResponsePatternAddedOrChangedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"p/a/id": {Name: "Inner", Properties: []string{"p/a/id", "q/a/id", "p/b/id"}, Count: 4},
	}, sharedAt(t, changes))
}

// A change below the shared schema, not in it, lists the other properties
// too, with the rest of its path appended.
func TestChangeBelowASharedSchemaListsTheOtherProperty(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_inner_change", checker.ResponsePropertyTypeChangedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/id": {Name: "Shared", Properties: []string{"left/id", "right/id"}, Count: 2},
	}, sharedAt(t, changes))
}

// The required list is Shared's own, so the change is at the properties that
// hold Shared, as in the response check.
func TestRequestPropertyBecameRequiredInASharedSchemaListsTheOtherProperty(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_request_required", checker.RequestPropertyRequiredUpdatedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left": {Name: "Shared", Properties: []string{"left", "right"}, Count: 2},
	}, sharedAt(t, changes))
}

// left and right reference Holder's inner property, so the schema they share
// is named after Holder, the component it is in.
func TestSharedSchemaReachedThroughARefIntoAComponentIsNamedAfterIt(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_ref_into_component", checker.ResponseOptionalPropertyUpdatedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"left/extra": {Name: "Holder", Properties: []string{"left/extra", "right/extra"}, Count: 2},
	}, sharedAt(t, changes))
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: `Holder`, also at `right/extra`)")
}

// A description beside each $ref to Address makes the parser copy Address
// once per property. The copies share their children, so zip is what both
// properties reach, and it is named after Address, the schema it belongs to.
func TestSharedSchemaWrittenInlineIsNamedAfterItsComponent(t *testing.T) {
	changes := sharedSchemaChanges(t, "shared_schema_override", checker.ResponsePatternAddedOrChangedCheck)

	require.Equal(t, map[string]*checker.SharedSchema{
		"billing/zip": {Name: "Address", Properties: []string{"billing/zip", "shipping/zip"}, Count: 2},
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
