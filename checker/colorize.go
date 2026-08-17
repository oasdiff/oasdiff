package checker

import (
	"fmt"

	"github.com/TwiN/go-color"

	"github.com/oasdiff/oasdiff/colorize"
)

// ColorMode is defined in the colorize package; the aliases keep the
// checker package's public surface unchanged.
type ColorMode = colorize.ColorMode

const (
	ColorAlways  = colorize.ColorAlways
	ColorNever   = colorize.ColorNever
	ColorAuto    = colorize.ColorAuto
	ColorInvalid = colorize.ColorInvalid
)

func GetSupportedColorValues() []string {
	return colorize.GetSupportedColorValues()
}

func NewColorMode(color string) (ColorMode, error) {
	return colorize.NewColorMode(color)
}

// IsColorEnabled lets oasdiff packages outside checker (validate, future
// subcommands) gate their own color logic on the same auto-detect +
// override convention.
func IsColorEnabled(colorMode ColorMode) bool {
	return colorize.IsColorEnabled(colorMode)
}

func isColorEnabled(colorMode ColorMode) bool {
	return colorize.IsColorEnabled(colorMode)
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

// SetPipedOutput overrides piped-output auto-detection; see colorize.
func SetPipedOutput(val *bool) *bool {
	return colorize.SetPipedOutput(val)
}
