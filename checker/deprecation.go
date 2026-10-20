package checker

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/diff"
)

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
