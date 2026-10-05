package consolidate_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/consolidate"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/stretchr/testify/require"
)

// checkFixture runs one check over a fixture pair under data/checker, named
// without its _base or _revision suffix.
func checkFixture(t *testing.T, fixture string, check checker.BackwardCompatibilityCheck) checker.Changes {
	t.Helper()
	base, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/checker/"+fixture+"_base.yaml"))
	require.NoError(t, err)
	revision, err := load.NewSpecInfo(openapi3.NewLoader(), load.NewSource("../data/checker/"+fixture+"_revision.yaml"))
	require.NoError(t, err)
	d, sources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), base, revision)
	require.NoError(t, err)
	config := checker.NewConfig(checker.BackwardCompatibilityChecks{check}, checker.WithSingleCheck(check), checker.WithSeverityLevels(map[string]checker.Level{
		checker.APIVersionNotBumpedId:      checker.NONE,
		checker.APIVersionDecreasedId:      checker.NONE,
		checker.APIMajorVersionNotBumpedId: checker.NONE,
	}))
	return checker.CheckBackwardCompatibilityUntilLevel(config, d, sources, checker.INFO)
}

// The pattern added to Id is reported at customerId and userId; merged, it
// is one finding at customerId that lists userId and explains itself.
func TestSharedSchema_MergesTheReferences(t *testing.T) {
	changes := consolidate.Changes(checkFixture(t, "shared_schema_two_properties", checker.ResponsePatternAddedOrChangedCheck), consolidate.SharedSchema)

	byOperation := map[string]checker.ApiChange{}
	for _, change := range changes {
		require.NotContains(t, byOperation, change.GetPath(), "%s reported twice", change.GetPath())
		byOperation[change.GetPath()] = change.(checker.ApiChange)
	}

	orders := byOperation["/orders"]
	require.Equal(t, "customerId", orders.Args[0])
	require.Equal(t, &checker.SharedSchema{Name: "Id", Properties: []string{"customerId", "userId"}}, orders.GetSharedSchema())
	require.Equal(t, checker.SharedSchemaCommentId, orders.Comment)
	require.Contains(t, orders.GetComment(checker.NewLocalizer("en")), "reported once per check")
	require.Contains(t, orders.GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: `Id`, also at `userId`)")

	accounts := byOperation["/accounts"]
	require.Equal(t, []string{"customerId", "ownerId", "userId"}, accounts.GetSharedSchema().Properties)
	require.Contains(t, accounts.GetUncolorizedText(checker.NewLocalizer("en")), "(shared schema: `Id`, also at `ownerId` and 1 more)")
}

// Each check merges its own findings: a property added to Shared and a
// pattern added inside it stay two findings, one per check.
func TestSharedSchema_PerCheck(t *testing.T) {
	added := checkFixture(t, "shared_schema_two_parents", checker.ResponseOptionalPropertyUpdatedCheck)
	pattern := checkFixture(t, "shared_schema_two_parents", checker.ResponsePatternAddedOrChangedCheck)

	ids := []string{}
	for _, change := range consolidate.Changes(append(added, pattern...), consolidate.SharedSchema) {
		if change.(checker.ApiChange).Comment == checker.SharedSchemaCommentId {
			ids = append(ids, change.GetId())
		}
	}
	require.ElementsMatch(t, []string{checker.ResponseOptionalPropertyAddedId, checker.ResponsePropertyPatternAddedId}, ids)
}

func sharedChange(property string, details string, pattern string) checker.ApiChange {
	others := map[string]string{"a": "b", "b": "a"}
	return checker.ApiChange{
		Id: "change_id", Operation: "GET", Path: "/x", Details: details,
		Args: []any{property, pattern},
	}.WithSharedSchema(&checker.SharedSchema{Name: "S", Properties: []string{property, others[property]}})
}

// Findings that differ in more than the property they are reported at are
// different changes, and are not merged.
func TestSharedSchema_KeepsDifferentChanges(t *testing.T) {
	changes := consolidate.SharedSchema(checker.Changes{
		sharedChange("a", "", "^x$"),
		sharedChange("b", "", "^x$"),
		sharedChange("a", "", "^y$"),
		sharedChange("a", "(media type: application/xml)", "^x$"),
	})

	require.Len(t, changes, 3)
	require.Equal(t, []any{"a", "^x$"}, changes[0].GetArgs())
	require.Equal(t, []any{"a", "^y$"}, changes[1].GetArgs())
	require.Equal(t, "(media type: application/xml)", changes[2].(checker.ApiChange).Details)
}

// A finding in no shared schema passes through as it is.
func TestSharedSchema_PassesOthersThrough(t *testing.T) {
	plain := checker.ApiChange{Id: "change_id", Args: []any{"a"}}
	require.Equal(t, checker.Changes{plain}, consolidate.SharedSchema(checker.Changes{plain}))
}

// The finding kept does not depend on the order the findings arrive in.
func TestSharedSchema_KeepsTheFirstPropertyInAnyOrder(t *testing.T) {
	changes := consolidate.SharedSchema(checker.Changes{sharedChange("b", "", "^x$"), sharedChange("a", "", "^x$")})

	require.Len(t, changes, 1)
	require.Equal(t, []any{"a", "^x$"}, changes[0].GetArgs())
}
