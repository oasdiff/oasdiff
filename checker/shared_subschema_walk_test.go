package checker_test

import (
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// A change to one referenced component is one contract change for an operation,
// regardless of how many properties provide paths to that component.
func TestResponseSharedSchemaChangeReportedOncePerOperation(t *testing.T) {
	base, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, checker.ResponseOptionalPropertyAddedId, changes[0].GetId())
}

func TestResponseSharedSchemaUsesStableRepresentativePath(t *testing.T) {
	base, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	for range 30 {
		d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
		require.NoError(t, err)
		changes := checker.CheckBackwardCompatibilityUntilLevel(
			singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
			d, sources, checker.INFO,
		)
		require.Len(t, changes, 1)
		change, ok := changes[0].(checker.ApiChange)
		require.True(t, ok)
		require.Equal(t, "left/extra", change.Args[0])
	}
}

// The change is reported once per schema, not once per parent: Shared is
// reached through First and through Second, which are different parents, and
// each of those through two properties of the payload.
func TestResponseSharedSchemaUnderTwoParentsReportedOnce(t *testing.T) {
	base, err := open("../data/checker/shared_schema_two_parents_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_two_parents_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)

	// The payload also adds properties of its own, which are not in Shared.
	inShared := checker.Changes{}
	for _, change := range changes {
		if strings.HasSuffix(change.(checker.ApiChange).Args[0].(string), "/shared/extra") {
			inShared = append(inShared, change)
		}
	}

	require.Len(t, inShared, 1)
	require.Equal(t, "first/shared/extra", inShared[0].(checker.ApiChange).Args[0])
	require.Equal(t, &checker.SharedSchema{Name: "Shared", Properties: []string{"first/shared/extra", "second/shared/extra"}}, inShared[0].(checker.ApiChange).GetSharedSchema())
}

// Every check walks the payload itself, so the deduplication is per check:
// Shared both gains a property and gains a pattern on an existing one, and
// each check reports its own change, both at the same property path.
func TestSharedSchemaIsReportedOncePerCheck(t *testing.T) {
	base, err := open("../data/checker/shared_schema_two_parents_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_two_parents_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, sources, checker.INFO)

	ids := []string{}
	for _, change := range changes {
		if change.(checker.ApiChange).Comment == checker.SharedSchemaCommentId {
			ids = append(ids, change.GetId())
			require.Contains(t, change.GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Shared, also at `second/shared/")
		}
	}

	require.ElementsMatch(t, []string{
		checker.ResponseOptionalPropertyAddedId,
		checker.ResponsePropertyPatternAddedId,
	}, ids)
}

// The pattern is added to Shared's own id property, so the change is below the
// schema several paths reach rather than in it, and names Shared all the same.
func TestChangeBelowASharedSchemaNamesIt(t *testing.T) {
	base, err := open("../data/checker/shared_schema_two_parents_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_two_parents_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, "first/shared/id", changes[0].(checker.ApiChange).Args[0])
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Shared, also at `second/shared/id`)")
}

// A change below a shared schema is reported at one path too, so it names the
// shared schema it is under: here the change is on Shared's own id property,
// and Shared is what several paths reach.
func TestChangeInsideASharedSchemaNamesTheSharedSchema(t *testing.T) {
	base, err := open("../data/checker/shared_schema_inner_change_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_inner_change_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponsePropertyTypeChangedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, "left/id", changes[0].(checker.ApiChange).Args[0])
	require.Equal(t, checker.SharedSchemaCommentId, changes[0].(checker.ApiChange).Comment)
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Shared, also at `right/id`)")
}

// A schema reached through a JSON pointer into another schema has no
// components.schemas name, so the change lists the other properties without
// one.
func TestSharedSchemaWithoutAComponentNameIsStillReportedOnce(t *testing.T) {
	base, err := open("../data/checker/shared_schema_unnamed_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_unnamed_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, checker.SharedSchemaCommentId, changes[0].(checker.ApiChange).Comment)
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(also at `right/extra`)")
	require.Empty(t, changes[0].(checker.ApiChange).GetSharedSchema().Name)
}

// A change in a schema several properties reach says so, and a change in a
// schema reached once does not.
func TestSharedSchemaChangeCarriesTheComment(t *testing.T) {
	sharedBase, err := open("../data/checker/shared_schema_base.yaml")
	require.NoError(t, err)
	sharedRevision, err := open("../data/checker/shared_schema_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), sharedBase, sharedRevision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)
	require.Len(t, changes, 1)
	require.Equal(t, checker.SharedSchemaCommentId, changes[0].(checker.ApiChange).Comment)
	require.NotEmpty(t, changes[0].GetComment(checker.NewLocalizer("en")))
	require.Contains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Shared, also at `right/extra`)")

	onceBase, err := open("../data/component-renamed1.yaml")
	require.NoError(t, err)
	onceRevision, err := open("../data/component-renamed2.yaml")
	require.NoError(t, err)

	d, sources, err = diff.GetWithOperationsSourcesMap(diff.NewConfig(), onceBase, onceRevision)
	require.NoError(t, err)
	changes = checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponseOptionalPropertyUpdatedCheck),
		d, sources, checker.INFO,
	)
	require.Len(t, changes, 1)
	require.Empty(t, changes[0].(checker.ApiChange).Comment)
	require.Nil(t, changes[0].(checker.ApiChange).GetSharedSchema())
	require.NotContains(t, changes[0].GetUncolorizedText(checker.NewLocalizer("en")), "also at")
}

// userId and customerId mean different things but share Id, so a pattern added
// to Id changes both. The change is reported once, at the property that sorts
// first, and names the other so a reviewer looking for userId finds it.
func TestSharedSchemaListsTheOtherProperties(t *testing.T) {
	base, err := open("../data/checker/shared_schema_two_properties_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_two_properties_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck),
		d, sources, checker.INFO,
	)
	require.Len(t, changes, 2)

	byPath := map[string]checker.ApiChange{}
	for _, change := range changes {
		byPath[change.GetPath()] = change.(checker.ApiChange)
	}

	orders := byPath["/orders"]
	require.Equal(t, &checker.SharedSchema{Name: "Id", Properties: []string{"customerId", "userId"}}, orders.GetSharedSchema())
	require.Contains(t, orders.GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Id, also at `userId`)")

	accounts := byPath["/accounts"]
	require.Equal(t, &checker.SharedSchema{Name: "Id", Properties: []string{"customerId", "ownerId", "userId"}}, accounts.GetSharedSchema())
	require.Contains(t, accounts.GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: Id, also at `ownerId` and 1 more)")
}

// A description beside each $ref to Address makes the parser copy Address
// once per property. The copies share their children, so zip is what both
// properties reach. zip is written inline and is named after Address, the
// schema it belongs to.
func TestSharedSchemaWrittenInlineIsNamedAfterItsComponent(t *testing.T) {
	base, err := open("../data/checker/shared_schema_override_base.yaml")
	require.NoError(t, err)
	revision, err := open("../data/checker/shared_schema_override_revision.yaml")
	require.NoError(t, err)

	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	changes := checker.CheckBackwardCompatibilityUntilLevel(
		singleCheckConfig(checker.ResponsePatternAddedOrChangedCheck),
		d, sources, checker.INFO,
	)

	require.Len(t, changes, 1)
	require.Equal(t, &checker.SharedSchema{Name: "Address", Properties: []string{"billing/zip", "shipping/zip"}}, changes[0].(checker.ApiChange).GetSharedSchema())
}
