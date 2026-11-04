package deprecation

import (
	"time"

	"cloud.google.com/go/civil"
)

// RemovalIds names the id that reports a removal at each stage of the
// deprecation process. SunsetParse is empty for a removed property, which has
// no id for an unparseable sunset date and reports Removed instead.
type RemovalIds struct {
	Removed         string
	WithDeprecation string
	BeforeSunset    string
	SunsetParse     string
}

// Removal follows the deprecation process for a removed element: its
// deprecation and sunset date decide which id reports it, and which arguments
// that id takes beyond the ones naming the element. ok is false once the
// sunset date has passed, where the removal is not reported at all.
func Removal(deprecated bool, extensions map[string]any, ids RemovalIds) (string, []any, bool) {
	if !deprecated {
		return ids.Removed, nil, true
	}

	sunset, found := sunsetExtension(extensions)
	if !found {
		return ids.WithDeprecation, nil, true
	}

	date, err := SunsetDate(sunset)
	if err != nil {
		if ids.SunsetParse == "" {
			return ids.Removed, nil, true
		}
		return ids.SunsetParse, []any{err}, true
	}

	if civil.DateOf(time.Now()).Before(date) {
		return ids.BeforeSunset, []any{date}, true
	}
	return "", nil, false
}
