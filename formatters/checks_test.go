package formatters_test

import (
	"slices"
	"testing"

	"github.com/oasdiff/oasdiff/formatters"
	"github.com/stretchr/testify/require"
)

func TestChecks_SortFunc(t *testing.T) {
	checks := formatters.Checks{
		{Id: "c", Level: "error", Description: "desc c"},
		{Id: "a", Level: "warn", Description: "desc a"},
		{Id: "b", Level: "info", Description: "desc b"},
	}

	slices.SortFunc(checks, checks.SortFunc)

	require.Equal(t, "a", checks[0].Id)
	require.Equal(t, "b", checks[1].Id)
	require.Equal(t, "c", checks[2].Id)
}

func TestChecks_SortFunc_Equal(t *testing.T) {
	checks := formatters.Checks{
		{Id: "a", Level: "error", Description: "desc 1"},
		{Id: "a", Level: "warn", Description: "desc 2"},
	}

	result := checks.SortFunc(checks[0], checks[1])
	require.Equal(t, 0, result)
}
