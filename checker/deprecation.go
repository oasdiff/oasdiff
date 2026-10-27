package checker

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"cloud.google.com/go/civil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// propertyRemovalId returns the rule id and arguments reporting a removed
// property, by its deprecation and sunset date as in checkParameterRemoval, or
// "" when the sunset date has passed. A sunset date that can't be parsed keeps
// removedId: the property deprecation checks report it.
func propertyRemovalId(property *openapi3.Schema, args []any, removedId, withDeprecationId, beforeSunsetId string) (string, []any) {
	if !property.Deprecated {
		return removedId, args
	}

	sunset, ok := getSunset(property.Extensions)
	if !ok {
		return withDeprecationId, args
	}

	date, err := getSunsetDate(sunset)
	if err != nil {
		return removedId, args
	}

	if civil.DateOf(time.Now()).Before(date) {
		return beforeSunsetId, append(args, date)
	}
	return "", nil
}

// sunsetChangeIds names the id that reports each change to a sunset date.
type sunsetChangeIds struct {
	deleted         string
	sunsetParse     string
	changedTooSmall string
}

// sunsetChangeId returns the id that reports a change to the sunset date of an
// element that stays deprecated, as RequestParameterSunsetChangedCheck does for
// parameters, and the arguments that id takes beyond the ones naming the
// element. ok is false when the change is not reported.
func sunsetChangeId(extensionsDiff *diff.ExtensionsDiff, base, revision map[string]any, deprecationDays uint, ids sunsetChangeIds) (string, []any, bool) {
	if extensionsDiff == nil {
		return "", nil, false
	}

	if slices.Contains(extensionsDiff.Deleted, diff.SunsetExtension) {
		return ids.deleted, nil, true
	}

	if _, ok := extensionsDiff.Modified[diff.SunsetExtension]; !ok {
		return "", nil, false
	}

	date, err := getSunsetDate(revision[diff.SunsetExtension])
	if err != nil {
		return ids.sunsetParse, []any{err}, true
	}

	// an invalid base date gave clients no date to rely on, so replacing it is a fix
	baseDate, err := getSunsetDate(base[diff.SunsetExtension])
	if err != nil {
		return "", nil, false
	}

	if baseDate.After(date) && date.DaysSince(civil.DateOf(time.Now())) < int(deprecationDays) {
		return ids.changedTooSmall, []any{baseDate, date, baseDate, deprecationDays}, true
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
