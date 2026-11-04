package checker

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

var deprecationTables = map[string]deprecationIds{
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
		for _, id := range []string{ids.reactivated, ids.deprecated, ids.withSunset, ids.sunsetMissing, ids.sunsetParse, ids.sunsetTooSmall} {
			if id == "" {
				continue
			}
			require.True(t, rules[id], "%s: %s", resource, id)
		}
	}

	// A parameter deprecated with a sunset date is reported as
	// request-parameter-deprecated, with the date in the details.
	require.Empty(t, requestParameterDeprecationIds.withSunset)
	for resource, ids := range deprecationTables {
		if resource != "parameter" {
			require.NotEmpty(t, ids.withSunset, resource)
		}
		require.NotEmpty(t, ids.reactivated, resource)
		require.NotEmpty(t, ids.deprecated, resource)
		require.NotEmpty(t, ids.sunsetMissing, resource)
		require.NotEmpty(t, ids.sunsetParse, resource)
		require.NotEmpty(t, ids.sunsetTooSmall, resource)
	}
}

func TestDeprecationIdStages(t *testing.T) {
	deprecating := &diff.ValueDiff{From: false, To: true}
	reactivating := &diff.ValueDiff{From: true, To: false}
	future := civil.DateOf(time.Now()).AddDays(200)
	ids := deprecationIds{
		reactivated:    "reactivated",
		deprecated:     "deprecated",
		withSunset:     "with-sunset",
		sunsetMissing:  "sunset-missing",
		sunsetParse:    "sunset-parse",
		sunsetTooSmall: "sunset-too-small",
	}

	for _, test := range []struct {
		name       string
		deprecated *diff.ValueDiff
		sunset     any
		days       uint
		expected   string
		atSunset   bool
	}{
		{name: "reactivated", deprecated: reactivating, expected: "reactivated"},
		{name: "reactivated with a grace period", deprecated: reactivating, days: 180, expected: "reactivated"},
		{name: "no sunset date and no grace period", deprecated: deprecating, expected: "deprecated"},
		{name: "no sunset date under a grace period", deprecated: deprecating, days: 180, expected: "sunset-missing"},
		{name: "unparseable sunset date", deprecated: deprecating, sunset: "not-a-date", expected: "sunset-parse", atSunset: true},
		{name: "sunset date inside the grace period", deprecated: deprecating, sunset: future.String(), days: 300, expected: "sunset-too-small", atSunset: true},
		{name: "sunset date beyond the grace period", deprecated: deprecating, sunset: future.String(), days: 180, expected: "with-sunset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			extensions := map[string]any{}
			if test.sunset != nil {
				extensions[diff.SunsetExtension] = test.sunset
			}

			stage := deprecationId(test.deprecated, extensions, test.days, ids)
			require.Equal(t, test.expected, stage.id)
			require.Equal(t, test.atSunset, stage.atSunset)
		})
	}

	// Without an id of its own, a valid sunset date is reported as deprecated,
	// and the date reaches the call site for the details.
	noWithSunset := ids
	noWithSunset.withSunset = ""
	stage := deprecationId(deprecating, map[string]any{diff.SunsetExtension: future.String()}, 180, noWithSunset)
	require.Equal(t, "deprecated", stage.id)
	require.Equal(t, future, stage.sunset)
}
