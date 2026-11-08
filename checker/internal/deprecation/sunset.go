package deprecation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

// InvalidSunsetDate reports a sunset extension value that is not an RFC3339 date.
type InvalidSunsetDate struct {
	// Value is the value the message reports, quoted when it is a string, so
	// that an empty or space-padded date is visible in every output format.
	Value string
}

func (e *InvalidSunsetDate) Error() string {
	return fmt.Sprintf("sunset date doesn't conform with RFC3339: %s", e.Value)
}

// sunsetValue is the message argument for an error SunsetDate returned: the
// value it could not parse, so that the localized message holds the sentence
// and the argument holds only the value.
func sunsetValue(err error) string {
	if invalid, ok := errors.AsType[*InvalidSunsetDate](err); ok {
		return invalid.Value
	}
	return err.Error()
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
		return civil.Date{}, &InvalidSunsetDate{Value: strconv.Quote(sunsetStr)}
	}

	return civil.Date{}, &InvalidSunsetDate{Value: extension.ValueText(sunset)}
}
