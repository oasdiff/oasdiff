package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func unevaluatedPropertiesChanges(t *testing.T, base, revision string) checker.Changes {
	t.Helper()
	s1, err := open("../data/checker/unevaluated_properties_" + base + ".yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/unevaluated_properties_" + revision + ".yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	errs := checker.CheckBackwardCompatibilityUntilLevel(allChecksConfig(), d, osm, checker.INFO)
	result := checker.Changes{}
	for _, e := range errs {
		switch e.GetId() {
		case checker.RequestBodyUnevaluatedPropertiesNarrowedId, checker.RequestBodyUnevaluatedPropertiesWidenedId,
			checker.RequestPropertyUnevaluatedPropertiesNarrowedId, checker.RequestPropertyUnevaluatedPropertiesWidenedId,
			checker.ResponseBodyUnevaluatedPropertiesNarrowedId, checker.ResponseBodyUnevaluatedPropertiesWidenedId,
			checker.ResponsePropertyUnevaluatedPropertiesNarrowedId, checker.ResponsePropertyUnevaluatedPropertiesWidenedId,
			checker.RequestBodyUnevaluatedPropertiesAddedId, checker.RequestBodyUnevaluatedPropertiesRemovedId,
			checker.RequestPropertyUnevaluatedPropertiesAddedId, checker.RequestPropertyUnevaluatedPropertiesRemovedId,
			checker.ResponseBodyUnevaluatedPropertiesAddedId, checker.ResponseBodyUnevaluatedPropertiesRemovedId,
			checker.ResponsePropertyUnevaluatedPropertiesAddedId, checker.ResponsePropertyUnevaluatedPropertiesRemovedId:
			result = append(result, e)
		}
	}
	return result
}

// closing an object (absent to false) and replacing a subschema with false both
// narrow, and the subschema disappearing is not also reported as removed
func TestUnevaluatedPropertiesNarrowed(t *testing.T) {
	changes := unevaluatedPropertiesChanges(t, "base", "revision")
	require.Equal(t, map[string]checker.Level{
		checker.RequestBodyUnevaluatedPropertiesNarrowedId:      checker.ERR,
		checker.RequestPropertyUnevaluatedPropertiesNarrowedId:  checker.ERR,
		checker.ResponseBodyUnevaluatedPropertiesNarrowedId:     checker.INFO,
		checker.ResponsePropertyUnevaluatedPropertiesNarrowedId: checker.INFO,
	}, changeSummary(changes))

	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyUnevaluatedPropertiesNarrowedId {
			require.Equal(t, []any{"details", "schema", "false"}, c.(checker.ApiChange).Args)
		}
		if c.GetId() == checker.ResponseBodyUnevaluatedPropertiesNarrowedId {
			require.Equal(t, []any{"true", "false", "200"}, c.(checker.ApiChange).Args)
		}
	}
}

// reopening a closed object and relaxing false to a subschema both widen, and
// the subschema appearing is not also reported as added
func TestUnevaluatedPropertiesWidened(t *testing.T) {
	require.Equal(t, map[string]checker.Level{
		checker.RequestBodyUnevaluatedPropertiesWidenedId:      checker.INFO,
		checker.RequestPropertyUnevaluatedPropertiesWidenedId:  checker.INFO,
		checker.ResponseBodyUnevaluatedPropertiesWidenedId:     checker.ERR,
		checker.ResponsePropertyUnevaluatedPropertiesWidenedId: checker.ERR,
	}, changeSummary(unevaluatedPropertiesChanges(t, "revision", "base")))
}

// a subschema appearing or disappearing on an open object keeps the
// unevaluated-properties-added/removed rules
func TestUnevaluatedPropertiesSubschemaAgainstOpen(t *testing.T) {
	require.Equal(t, map[string]checker.Level{
		checker.RequestPropertyUnevaluatedPropertiesRemovedId:  checker.INFO,
		checker.ResponsePropertyUnevaluatedPropertiesRemovedId: checker.ERR,
	}, changeSummary(unevaluatedPropertiesChanges(t, "base", "true")))
	require.Equal(t, map[string]checker.Level{
		checker.RequestPropertyUnevaluatedPropertiesAddedId:  checker.ERR,
		checker.ResponsePropertyUnevaluatedPropertiesAddedId: checker.INFO,
	}, changeSummary(unevaluatedPropertiesChanges(t, "true", "base")))
}

// true and absent are the same open form
func TestUnevaluatedPropertiesTrueToFalse(t *testing.T) {
	require.Equal(t, map[string]checker.Level{
		checker.RequestBodyUnevaluatedPropertiesNarrowedId:      checker.ERR,
		checker.RequestPropertyUnevaluatedPropertiesNarrowedId:  checker.ERR,
		checker.ResponseBodyUnevaluatedPropertiesNarrowedId:     checker.INFO,
		checker.ResponsePropertyUnevaluatedPropertiesNarrowedId: checker.INFO,
	}, changeSummary(unevaluatedPropertiesChanges(t, "true", "revision")))
}
