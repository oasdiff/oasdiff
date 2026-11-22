package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func additionalPropertiesChanges(t *testing.T, base, revision string) checker.Changes {
	t.Helper()
	s1, err := open("../data/checker/additional_properties_" + base + ".yaml")
	require.NoError(t, err)
	s2, err := open("../data/checker/additional_properties_" + revision + ".yaml")
	require.NoError(t, err)
	d, osm, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), s1, s2)
	require.NoError(t, err)
	config := allChecksConfig()
	errs := checker.CheckBackwardCompatibilityUntilLevel(config, d, osm, checker.INFO)
	result := checker.Changes{}
	for _, e := range errs {
		switch e.GetId() {
		case checker.RequestBodyAdditionalPropertiesNarrowedId, checker.RequestBodyAdditionalPropertiesWidenedId,
			checker.RequestPropertyAdditionalPropertiesNarrowedId, checker.RequestPropertyAdditionalPropertiesWidenedId,
			checker.ResponseBodyAdditionalPropertiesNarrowedId, checker.ResponseBodyAdditionalPropertiesWidenedId,
			checker.ResponsePropertyAdditionalPropertiesNarrowedId, checker.ResponsePropertyAdditionalPropertiesWidenedId:
			result = append(result, e)
		}
	}
	return result
}

func changeSummary(changes checker.Changes) map[string]checker.Level {
	result := map[string]checker.Level{}
	for _, c := range changes {
		result[c.GetId()] = c.GetLevel()
	}
	return result
}

// closing an object (absent to false) and replacing a subschema with false both narrow
func TestAdditionalPropertiesNarrowed(t *testing.T) {
	changes := additionalPropertiesChanges(t, "base", "revision")
	require.Equal(t, map[string]checker.Level{
		checker.RequestBodyAdditionalPropertiesNarrowedId:      checker.ERR,
		checker.RequestPropertyAdditionalPropertiesNarrowedId:  checker.ERR,
		checker.ResponseBodyAdditionalPropertiesNarrowedId:     checker.INFO,
		checker.ResponsePropertyAdditionalPropertiesNarrowedId: checker.INFO,
	}, changeSummary(changes))

	for _, c := range changes {
		if c.GetId() == checker.RequestPropertyAdditionalPropertiesNarrowedId {
			require.Equal(t, []any{"details", "schema", "false"}, c.(checker.ApiChange).Args)
		}
		if c.GetId() == checker.ResponseBodyAdditionalPropertiesNarrowedId {
			require.Equal(t, []any{"true", "false", "200"}, c.(checker.ApiChange).Args)
		}
	}
}

// reopening a closed object and relaxing false to a subschema both widen
func TestAdditionalPropertiesWidened(t *testing.T) {
	require.Equal(t, map[string]checker.Level{
		checker.RequestBodyAdditionalPropertiesWidenedId:      checker.INFO,
		checker.RequestPropertyAdditionalPropertiesWidenedId:  checker.INFO,
		checker.ResponseBodyAdditionalPropertiesWidenedId:     checker.ERR,
		checker.ResponsePropertyAdditionalPropertiesWidenedId: checker.ERR,
	}, changeSummary(additionalPropertiesChanges(t, "revision", "base")))
}

// absent and true admit the same keys, so switching between them is not a change
func TestAdditionalPropertiesAbsentToTrue(t *testing.T) {
	require.Empty(t, additionalPropertiesChanges(t, "base", "true"))
}
