package checker

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/stretchr/testify/require"
)

var deprecationTables = map[string]deprecation.Ids{
	"endpoint":          apiDeprecationIds,
	"parameter":         requestParameterDeprecationIds,
	"request property":  requestPropertyDeprecationIds,
	"response property": responsePropertyDeprecationIds,
}

// Every id the deprecation tables name must be a registered rule, or the
// change it reports has no message and no level of its own.
func TestDeprecationIdsAreRules(t *testing.T) {
	rules := map[string]bool{}
	for _, rule := range GetAllRules() {
		rules[rule.Id] = true
	}

	for resource, ids := range deprecationTables {
		for _, id := range []string{ids.Reactivated, ids.Deprecated, ids.WithSunset, ids.SunsetMissing, ids.SunsetParse, ids.SunsetTooSmall} {
			if id == "" {
				continue
			}
			require.True(t, rules[id], "%s: %s", resource, id)
		}
	}

	// A parameter deprecated with a sunset date is reported as
	// request-parameter-deprecated, with the date in the details.
	require.Empty(t, requestParameterDeprecationIds.WithSunset)
	for resource, ids := range deprecationTables {
		if resource != "parameter" {
			require.NotEmpty(t, ids.WithSunset, resource)
		}
		require.NotEmpty(t, ids.Reactivated, resource)
		require.NotEmpty(t, ids.Deprecated, resource)
		require.NotEmpty(t, ids.SunsetMissing, resource)
		require.NotEmpty(t, ids.SunsetParse, resource)
		require.NotEmpty(t, ids.SunsetTooSmall, resource)
	}
}
