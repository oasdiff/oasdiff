package checker

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/oasdiff/oasdiff/diff"
	"github.com/stretchr/testify/require"
)

// bytes that are not valid json report the bytes. The diff marshals an
// extension before any check runs, so this value cannot arrive from a spec,
// only from a caller that builds the extensions itself.
func TestGetStabilityLevel_InvalidJson(t *testing.T) {
	_, err := getStabilityLevel(map[string]any{diff.XStabilityLevelExtension: json.RawMessage(`{"unterminated":`)})
	invalid, ok := errors.AsType[*InvalidStabilityLevel](err)
	require.True(t, ok, "expected an InvalidStabilityLevel, got %v", err)
	require.Equal(t, `"{\"unterminated\":"`, invalid.Value)
}
