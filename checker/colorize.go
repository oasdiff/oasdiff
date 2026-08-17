package checker

import (
	"fmt"

	"github.com/TwiN/go-color"

	"github.com/oasdiff/oasdiff/checker/rules"
)

// ColorMode is defined in checker/rules; the aliases keep the checker
// package's public surface unchanged.
type ColorMode = rules.ColorMode

const (
	ColorAlways  = rules.ColorAlways
	ColorNever   = rules.ColorNever
	ColorAuto    = rules.ColorAuto
	ColorInvalid = rules.ColorInvalid
)

func GetSupportedColorValues() []string {
	return rules.GetSupportedColorValues()
}

func NewColorMode(color string) (ColorMode, error) {
	return rules.NewColorMode(color)
}

// IsColorEnabled lets oasdiff packages outside checker (validate, future
// subcommands) gate their own color logic on the same auto-detect +
// override convention.
func IsColorEnabled(colorMode ColorMode) bool {
	return rules.IsColorEnabled(colorMode)
}

func isColorEnabled(colorMode ColorMode) bool {
	return rules.IsColorEnabled(colorMode)
}

func colorizedValues(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		result[i] = color.InBold(fmt.Sprintf("'%s'", interfaceToString(arg)))
	}
	return result
}

func quotedValues(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		result[i] = fmt.Sprintf("`%s`", interfaceToString(arg))
	}
	return result
}

// SetPipedOutput overrides piped-output auto-detection; see checker/rules.
func SetPipedOutput(val *bool) *bool {
	return rules.SetPipedOutput(val)
}
