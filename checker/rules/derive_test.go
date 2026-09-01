package rules_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/stretchr/testify/require"
)

// The effect and direction cells of the law, with no guards: narrowing
// breaks request consumers, widening breaks response consumers, an
// incomparable change breaks both, an unknown one warns, and a violation is
// an error on either side.
func TestDeriveLevel(t *testing.T) {
	for _, tc := range []struct {
		effect    rules.Effect
		direction rules.Direction
		level     rules.Level
	}{
		{rules.EffectNarrows, rules.DirectionRequest, rules.ERR},
		{rules.EffectNarrows, rules.DirectionResponse, rules.INFO},
		{rules.EffectWidens, rules.DirectionRequest, rules.INFO},
		{rules.EffectWidens, rules.DirectionResponse, rules.ERR},
		{rules.EffectIncomparable, rules.DirectionRequest, rules.ERR},
		{rules.EffectIncomparable, rules.DirectionResponse, rules.ERR},
		{rules.EffectViolation, rules.DirectionRequest, rules.ERR},
		{rules.EffectViolation, rules.DirectionResponse, rules.ERR},
		{rules.EffectUnknown, rules.DirectionRequest, rules.WARN},
		{rules.EffectUnknown, rules.DirectionResponse, rules.WARN},
		{rules.EffectNone, rules.DirectionRequest, rules.INFO},
		{rules.EffectNone, rules.DirectionResponse, rules.INFO},
		{rules.EffectNarrows, rules.DirectionNone, rules.ERR},
		{rules.EffectWidens, rules.DirectionNone, rules.INFO},
	} {
		require.Equal(t, tc.level, rules.DeriveLevel(tc.effect, tc.direction),
			"effect=%s direction=%s", tc.effect, tc.direction)
	}
}

// A guard nullifies or requalifies the effect only on the side it speaks
// about: readOnly removes a property from requests, writeOnly from
// responses, and on the other side each leaves the verdict alone.
func TestDeriveLevelGuards(t *testing.T) {
	for _, tc := range []struct {
		name      string
		effect    rules.Effect
		direction rules.Direction
		guard     rules.Guard
		level     rules.Level
	}{
		{"readOnly nullifies a request narrowing", rules.EffectNarrows, rules.DirectionRequest, rules.GuardReadOnly, rules.INFO},
		{"readOnly leaves a response widening breaking", rules.EffectWidens, rules.DirectionResponse, rules.GuardReadOnly, rules.ERR},
		{"writeOnly nullifies a response widening", rules.EffectWidens, rules.DirectionResponse, rules.GuardWriteOnly, rules.INFO},
		{"writeOnly leaves a request narrowing breaking", rules.EffectNarrows, rules.DirectionRequest, rules.GuardWriteOnly, rules.ERR},
		{"non-success nullifies on either side", rules.EffectWidens, rules.DirectionResponse, rules.GuardNonSuccess, rules.INFO},
		{"sanctioned is info even for a violation", rules.EffectViolation, rules.DirectionRequest, rules.GuardSanctioned, rules.INFO},
		{"negotiated gives a response narrowing request polarity", rules.EffectNarrows, rules.DirectionResponse, rules.GuardNegotiated, rules.ERR},
		{"negotiated makes a response widening safe", rules.EffectWidens, rules.DirectionResponse, rules.GuardNegotiated, rules.INFO},
	} {
		require.Equal(t, tc.level, rules.DeriveLevel(tc.effect, tc.direction, tc.guard), tc.name)
	}
}
