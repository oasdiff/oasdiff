package deprecation

import (
	"slices"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

// SunsetChangeIds names the id that reports each change to a sunset date.
type SunsetChangeIds struct {
	Deleted         string
	SunsetParse     string
	ChangedTooSmall string
}

// SunsetChange follows a change to the sunset date of an element that stays
// deprecated: the change and the sunset grace period decide which id reports
// it, and which arguments that id takes beyond the ones naming the element.
// daysKnown is false when the stability level does not resolve to a grace
// period. ok is false when the change is not reported.
func SunsetChange(extensionsDiff *diff.ExtensionsDiff, base, revision map[string]any, days uint, daysKnown bool, ids SunsetChangeIds) (string, []any, bool) {
	if extensionsDiff == nil {
		return "", nil, false
	}

	if slices.Contains(extensionsDiff.Deleted, diff.SunsetExtension) {
		return ids.Deleted, nil, true
	}

	if _, ok := extensionsDiff.Modified[diff.SunsetExtension]; !ok {
		return "", nil, false
	}

	date, err := SunsetDate(revision[diff.SunsetExtension])
	if err != nil {
		return ids.SunsetParse, []any{err}, true
	}

	baseDate, err := SunsetDate(base[diff.SunsetExtension])
	if err != nil {
		// an unparseable base date gave clients no date to rely on, so replacing it is not reported
		return "", nil, false
	}

	if !daysKnown {
		// the grace period is what the comparison below needs, and the cases
		// above do not, so there is nothing left to report
		return "", nil, false
	}

	if baseDate.After(date) && date.DaysSince(civil.DateOf(time.Now())) < int(days) {
		return ids.ChangedTooSmall, []any{baseDate, date, baseDate, days}, true
	}
	return "", nil, false
}
