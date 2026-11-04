package checker

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

// removalIds names the id that reports a removal at each stage of the
// deprecation process. sunsetParse is empty where another check already
// reports an unparseable sunset date, and the removal reports removed instead.
type removalIds struct {
	removed         string
	withDeprecation string
	beforeSunset    string
	sunsetParse     string
}

// removalId follows the deprecation process for a removed element: its
// deprecation and sunset date decide which id reports it, and which arguments
// that id takes beyond the ones naming the element. ok is false once the
// sunset date has passed, where the removal is not reported at all.
func removalId(deprecated bool, extensions map[string]any, ids removalIds) (string, []any, bool) {
	if !deprecated {
		return ids.removed, nil, true
	}

	sunset, found := getSunset(extensions)
	if !found {
		return ids.withDeprecation, nil, true
	}

	date, err := getSunsetDate(sunset)
	if err != nil {
		if ids.sunsetParse == "" {
			return ids.removed, nil, true
		}
		return ids.sunsetParse, []any{err}, true
	}

	if civil.DateOf(time.Now()).Before(date) {
		return ids.beforeSunset, []any{date}, true
	}
	return "", nil, false
}

func getSunset(Extensions map[string]any) (any, bool) {
	value, ok := Extensions[diff.SunsetExtension]
	return value, ok
}

func getSunsetDate(sunset any) (civil.Date, error) {
	if sunsetJson, ok := sunset.(json.RawMessage); ok {
		if err := json.Unmarshal(sunsetJson, &sunset); err != nil {
			return civil.Date{}, fmt.Errorf("failed to unmarshal sunset json: %w", err)
		}
	}

	if sunsetStr, ok := sunset.(string); ok {
		if date, err := civil.ParseDate(sunsetStr); err == nil {
			return date, nil
		}
		if date, err := time.Parse(time.RFC3339, sunsetStr); err == nil {
			return civil.DateOf(date), nil
		}
	}

	return civil.Date{}, fmt.Errorf("sunset date doesn't conform with RFC3339: %s", extensionValueText(sunset))
}
