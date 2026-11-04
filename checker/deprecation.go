package checker

import (
	"encoding/json"
	"fmt"
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

// deprecationIds names the id that reports each stage of announcing a
// deprecation. withSunset is empty where the resource reports deprecated with
// the sunset date in its details instead of under an id of its own.
type deprecationIds struct {
	reactivated    string
	deprecated     string
	withSunset     string
	sunsetMissing  string
	sunsetParse    string
	sunsetTooSmall string
}

// deprecationStage is the stage of the deprecation process that an element's
// changed deprecated flag puts it in.
type deprecationStage struct {
	id       string
	args     []any      // the arguments the id takes beyond the ones naming the element
	sunset   civil.Date // zero where the element has no sunset date, or it doesn't parse
	atSunset bool       // the finding is about the sunset date rather than the flag
}

// nowDeprecated reports whether a changed deprecated flag is deprecating the
// element rather than reactivating it.
func nowDeprecated(deprecatedDiff *diff.ValueDiff) bool {
	return deprecatedDiff.To != nil && deprecatedDiff.To != false
}

// deprecationId follows the deprecation of an element whose deprecated flag
// changed: the flag, the sunset date and the sunset grace period decide which
// id reports it, and which arguments that id takes beyond the ones naming the
// element.
func deprecationId(deprecatedDiff *diff.ValueDiff, extensions map[string]any, days uint, ids deprecationIds) deprecationStage {
	if !nowDeprecated(deprecatedDiff) {
		return deprecationStage{id: ids.reactivated}
	}

	sunset, found := getSunset(extensions)
	if !found {
		if days > 0 {
			return deprecationStage{id: ids.sunsetMissing}
		}
		return deprecationStage{id: ids.deprecated}
	}

	date, err := getSunsetDate(sunset)
	if err != nil {
		return deprecationStage{id: ids.sunsetParse, args: []any{err}, atSunset: true}
	}

	if date.DaysSince(civil.DateOf(time.Now())) < int(days) {
		return deprecationStage{id: ids.sunsetTooSmall, args: []any{date, days}, sunset: date, atSunset: true}
	}

	if ids.withSunset == "" {
		return deprecationStage{id: ids.deprecated, sunset: date}
	}
	return deprecationStage{id: ids.withSunset, args: []any{date}, sunset: date}
}

// deprecationDays is the sunset grace period for an operation. ok is false
// where its stability level doesn't parse.
func deprecationDays(config *Config, extensions map[string]any) (days uint, ok bool) {
	stability, err := getStabilityLevel(extensions)
	if err != nil {
		return 0, false
	}
	return getDeprecationDays(config, stability), true
}
