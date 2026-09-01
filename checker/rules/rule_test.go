package rules_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/stretchr/testify/require"
)

func TestDerivedLevel(t *testing.T) {
	r := rules.Rule{
		Effect:    rules.EffectNarrows,
		Direction: rules.DirectionRequest,
		Guards:    []rules.Guard{rules.GuardReadOnly},
	}
	require.Equal(t, rules.INFO, r.DerivedLevel())
}
