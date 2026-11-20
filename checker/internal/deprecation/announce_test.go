package deprecation_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

func TestAnnounceStages(t *testing.T) {
	deprecating := &diff.ValueDiff{From: false, To: true}
	reactivating := &diff.ValueDiff{From: true, To: false}
	future := civil.DateOf(time.Now()).AddDays(200)
	ids := deprecation.Ids{
		Reactivated:    "reactivated",
		Deprecated:     "deprecated",
		WithSunset:     "with-sunset",
		SunsetMissing:  "sunset-missing",
		SunsetParse:    "sunset-parse",
		SunsetTooSmall: "sunset-too-small",
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

			stage := deprecation.Announce(test.deprecated, extensions, test.days, ids)
			require.Equal(t, test.expected, stage.Id)
			require.Equal(t, test.atSunset, stage.AtSunset)
		})
	}

	// Without an id of its own, a valid sunset date is reported as deprecated,
	// and the date reaches the call site for the details.
	noWithSunset := ids
	noWithSunset.WithSunset = ""
	stage := deprecation.Announce(deprecating, map[string]any{diff.SunsetExtension: future.String()}, 180, noWithSunset)
	require.Equal(t, "deprecated", stage.Id)
	require.Equal(t, future, stage.Sunset)
}
