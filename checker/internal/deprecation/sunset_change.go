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
	ChangedEarlier  string
	ChangedTooSmall string
	// ChangedEarlierComment explains why withdrawing a published date is
	// reported even when no deprecation policy is configured.
	ChangedEarlierComment string
}

// SunsetChange follows a change to the sunset date of an element that stays
// deprecated: the change and the sunset grace period decide which id reports
// it, and which arguments that id takes beyond the ones naming the element.
// daysKnown is false when the stability level does not resolve to a grace
// period. ok is false when the change is not reported.
func SunsetChange(extensionsDiff *diff.ExtensionsDiff, base, revision map[string]any, days uint, daysKnown bool, ids SunsetChangeIds) (string, []any, string, bool) {
	if extensionsDiff == nil {
		return "", nil, "", false
	}

	if slices.Contains(extensionsDiff.Deleted, diff.SunsetExtension) {
		return ids.Deleted, nil, "", true
	}

	if _, ok := extensionsDiff.Modified[diff.SunsetExtension]; !ok {
		return "", nil, "", false
	}

	date, err := SunsetDate(revision[diff.SunsetExtension])
	if err != nil {
		return ids.SunsetParse, []any{sunsetValue(err)}, "", true
	}

	baseDate, err := SunsetDate(base[diff.SunsetExtension])
	if err != nil {
		// an unparseable base date gave clients no date to rely on, so replacing it is not reported
		return "", nil, "", false
	}

	if !baseDate.After(date) {
		return "", nil, "", false
	}

	// the date moved earlier than the one clients were given, which is the
	// breaking change; the grace period only decides which id reports it, so a
	// stability level that does not resolve to one still reports the move
	if daysKnown && date.DaysSince(civil.DateOf(time.Now())) < int(days) {
		return ids.ChangedTooSmall, []any{baseDate, date, baseDate, days}, "", true
	}
	return ids.ChangedEarlier, []any{baseDate, date}, ids.ChangedEarlierComment, true
}
