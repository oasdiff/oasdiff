package internal_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/internal"
	"github.com/stretchr/testify/require"
)

// The --fail-on flag and the config file accept these names, and the run
// then converts them with the checker: a name the checker does not accept
// would reach the user as a run-time error on a value oasdiff had already
// told them was valid.
func TestLevels_AcceptedByChecker(t *testing.T) {
	for _, level := range internal.GetSupportedLevels() {
		_, err := checker.NewLevel(level)
		require.NoError(t, err, level)
	}
}
