package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/stretchr/testify/require"
)

const (
	numOfChecks = 95
	numOfIds    = 277
)

func TestNewConfig(t *testing.T) {
	config := allChecksConfig()
	require.Len(t, config.Checks, numOfChecks)
	require.Len(t, config.LogLevels, numOfIds)
	require.Equal(t, checker.DefaultBetaDeprecationDays, config.MinSunsetBetaDays)
	require.Equal(t, checker.DefaultStableDeprecationDays, config.MinSunsetStableDays)
}

func TestNewConfigWithDeprecation(t *testing.T) {
	config := allChecksConfig().WithDeprecation(10, 20)
	require.Len(t, config.Checks, numOfChecks)
	require.Len(t, config.LogLevels, numOfIds)
	require.Equal(t, uint(10), config.MinSunsetBetaDays)
	require.Equal(t, uint(20), config.MinSunsetStableDays)
}

func TestNewConfigWithOptionalCheck(t *testing.T) {
	const id = checker.RequestPropertyDefaultValueChangedId
	config := allChecksConfig().WithOptionalCheck(id)
	require.Equal(t, checker.ERR, config.LogLevels[id])
}

func TestNewConfigWithSeverityLevels(t *testing.T) {
	const id = checker.RequestPropertyDefaultValueChangedId
	config := allChecksConfig().WithSeverityLevels(map[string]checker.Level{id: checker.ERR})
	require.Equal(t, checker.ERR, config.LogLevels[id])
}
