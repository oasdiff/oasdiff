package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

func TestGetOptionalRuleIds(t *testing.T) {
	require.Len(t, checker.GetOptionalRuleIds(), 10)
}
