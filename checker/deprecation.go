package checker

import (
	"encoding/json"
	"errors"
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
