package checker

import (
	"encoding/json"
	"errors"
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
	sunsetStr, ok := sunset.(string)
	if !ok {
		sunsetJson, ok := sunset.(json.RawMessage)
		if !ok {
			return civil.Date{}, errors.New("sunset value isn't a string nor valid json")
		}
		if err := json.Unmarshal(sunsetJson, &sunsetStr); err != nil {
			return civil.Date{}, fmt.Errorf("failed to unmarshal sunset json: %w", err)
		}
	}

	if date, err := civil.ParseDate(sunsetStr); err == nil {
		return date, nil
	} else if date, err := time.Parse(time.RFC3339, sunsetStr); err == nil {
		return civil.DateOf(date), nil
	}

	return civil.Date{}, fmt.Errorf("sunset date doesn't conform with RFC3339: %s", sunsetStr)
}
