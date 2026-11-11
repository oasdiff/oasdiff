package deprecation

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

// Ids names the id that reports each stage of announcing a deprecation.
// WithSunset is empty where the resource reports deprecated with the sunset
// date in its details instead of under an id of its own.
type Ids struct {
	Reactivated    string
	Deprecated     string
	WithSunset     string
	SunsetMissing  string
	SunsetParse    string
	SunsetTooSmall string
}

// Stage is the stage of the deprecation process that an element's changed
// deprecated flag puts it in.
type Stage struct {
	Id       string
	Args     []any      // the arguments the id takes beyond the ones naming the element
	Sunset   civil.Date // zero where the element has no sunset date, or it doesn't parse
	AtSunset bool       // the finding is about the sunset date rather than the flag
}

// NowDeprecated reports whether a changed deprecated flag is deprecating the
// element rather than reactivating it.
func NowDeprecated(deprecatedDiff *diff.ValueDiff) bool {
	return deprecatedDiff.To != nil && deprecatedDiff.To != false
}

// Announce follows the deprecation of an element whose deprecated flag
// changed: the flag, the sunset date and the sunset grace period decide which
// id reports it, and which arguments that id takes beyond the ones naming the
// element.
func Announce(deprecatedDiff *diff.ValueDiff, extensions map[string]any, days uint, ids Ids) Stage {
	if !NowDeprecated(deprecatedDiff) {
		return Stage{Id: ids.Reactivated}
	}

	sunset, found := sunsetExtension(extensions)
	if !found {
		if days > 0 {
			return Stage{Id: ids.SunsetMissing}
		}
		return Stage{Id: ids.Deprecated}
	}

	date, err := SunsetDate(sunset)
	if err != nil {
		return Stage{Id: ids.SunsetParse, Args: []any{sunsetValue(err)}, AtSunset: true}
	}

	if date.DaysSince(civil.DateOf(time.Now())) < int(days) {
		return Stage{Id: ids.SunsetTooSmall, Args: []any{date, days}, Sunset: date, AtSunset: true}
	}

	if ids.WithSunset == "" {
		return Stage{Id: ids.Deprecated, Sunset: date}
	}
	return Stage{Id: ids.WithSunset, Args: []any{date}, Sunset: date}
}
