package deprecation_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/oasdiff/oasdiff/checker/internal/deprecation"
	"github.com/stretchr/testify/require"
)

// the value a message reports is quoted when it is a string, so that a value
// which is empty, or a valid date padded with spaces, is visible
func TestSunsetDate_InvalidValue(t *testing.T) {
	for _, test := range []struct {
		value any
		want  string
	}{
		{value: "not-a-date", want: `"not-a-date"`},
		{value: "", want: `""`},
		{value: " ", want: `" "`},
		{value: "2099-12-31 ", want: `"2099-12-31 "`},
		{value: json.RawMessage(`"2099-12-31 "`), want: `"2099-12-31 "`},
		{value: nil, want: "null"},
		{value: json.RawMessage(`{"invalid":"object"}`), want: `{"invalid":"object"}`},
	} {
		_, err := deprecation.SunsetDate(test.value)
		invalid, ok := errors.AsType[*deprecation.InvalidSunsetDate](err)
		require.True(t, ok, "expected an InvalidSunsetDate, got %v", err)
		require.Equal(t, test.want, invalid.Value)
	}
}

// a caller that prints the error, rather than reporting a change, gets a sentence
func TestSunsetDate_ErrorText(t *testing.T) {
	_, err := deprecation.SunsetDate("not-a-date")
	require.EqualError(t, err, `sunset date doesn't conform with RFC3339: "not-a-date"`)
}

// a date that parses reports no error, whether written as a plain date or with a time
func TestSunsetDate_Parsed(t *testing.T) {
	for _, value := range []any{"2099-12-31", "2099-12-31T00:00:00Z", json.RawMessage(`"2099-12-31"`)} {
		date, err := deprecation.SunsetDate(value)
		require.NoError(t, err)
		require.Equal(t, "2099-12-31", date.String())
	}
}

// bytes that are not valid json report the bytes, not a json error
func TestSunsetDate_InvalidJson(t *testing.T) {
	_, err := deprecation.SunsetDate(json.RawMessage(`{"unterminated":`))
	invalid, ok := errors.AsType[*deprecation.InvalidSunsetDate](err)
	require.True(t, ok, "expected an InvalidSunsetDate, got %v", err)
	require.Equal(t, `"{\"unterminated\":"`, invalid.Value)
}
