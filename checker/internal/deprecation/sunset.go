package deprecation

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/oasdiff/oasdiff/checker/internal/extension"
	"github.com/oasdiff/oasdiff/diff"
)

// sunsetExtension is the element's x-sunset extension, if it has one.
func sunsetExtension(extensions map[string]any) (any, bool) {
	value, ok := extensions[diff.SunsetExtension]
	return value, ok
}

// SunsetDate parses a sunset extension value as an RFC3339 date.
func SunsetDate(sunset any) (civil.Date, error) {
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

	return civil.Date{}, fmt.Errorf("sunset date doesn't conform with RFC3339: %s", extension.ValueText(sunset))
}
