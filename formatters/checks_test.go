package formatters_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

func TestChecks_SortFunc(t *testing.T) {
	checks := formatters.Checks{
		{Id: "c", Level: "error", Description: "desc c"},
		{Id: "a", Level: "warn", Description: "desc a"},
		{Id: "b", Level: "info", Description: "desc b"},
	}

	slices.SortFunc(checks, checks.SortFunc)

	require.Equal(t, "a", checks[0].Id)
	require.Equal(t, "b", checks[1].Id)
	require.Equal(t, "c", checks[2].Id)
}

func TestChecks_SortFunc_Equal(t *testing.T) {
	checks := formatters.Checks{
		{Id: "a", Level: "error", Description: "desc 1"},
		{Id: "a", Level: "warn", Description: "desc 2"},
	}

	result := checks.SortFunc(checks[0], checks[1])
	require.Equal(t, 0, result)
}

// Check's fields are omitempty so the validate listing, which has no direction,
// area, kind or action, doesn't render rows of empty strings. That is only safe
// while the changelog and breaking-change rules populate every field: an empty
// one would silently disappear from json/yaml rather than show as "".
//
// This pins the invariant at the source, so a new rule missing an area, kind,
// action or description fails here instead of quietly changing machine output.
func TestChecks_AllRuleFieldsPopulated(t *testing.T) {
	localizer := checker.NewLocalizer("en")

	for _, rule := range checker.GetAllRules() {
		require.NotEmpty(t, rule.Id, "rule id")
		require.NotEmpty(t, rule.Level.String(), "level for %s", rule.Id)
		require.NotEmpty(t, rule.Direction.String(), "direction for %s", rule.Id)
		require.NotEmpty(t, rule.Area.String(), "area for %s", rule.Id)
		require.NotEmpty(t, rule.Kind.String(), "kind for %s", rule.Id)
		require.NotEmpty(t, rule.Actions(), "actions for %s", rule.Id)
		require.NotEmpty(t, rule.Effect.String(), "effect for %s", rule.Id)
		require.NotEmpty(t, localizer(rule.Description), "description for %s", rule.Id)
	}
}

// A fully populated check keeps every field: omitempty must not drop anything
// the changelog listing reports today.
func TestChecks_FullCheckRendersAllFields(t *testing.T) {
	checks := formatters.Checks{{
		Id: "some-rule", Level: "error", Direction: "request", Area: "schema",
		Kind: "type", Actions: []string{"change"}, Effect: "widens", Description: "d", Mitigation: "m",
	}}

	out, err := formatters.JSONFormatter{}.RenderChecks(checks, formatters.NewRenderOpts())
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	require.Len(t, got, 1)
	for _, field := range []string{"id", "level", "direction", "area", "kind", "actions", "effect", "description", "mitigation"} {
		require.Contains(t, got[0], field, "field %s must survive omitempty when populated", field)
	}
}

// An IDs-only check renders as just the id, which is the point of omitempty.
func TestChecks_IdOnlyCheckOmitsEmptyFields(t *testing.T) {
	out, err := formatters.JSONFormatter{}.RenderChecks(formatters.Checks{{Id: "some-rule"}}, formatters.NewRenderOpts())
	require.NoError(t, err)

	var got []map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	require.Equal(t, []map[string]any{{"id": "some-rule"}}, got)
}
