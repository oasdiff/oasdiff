package checker

import (
	"fmt"

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

// ArgFormat sets a message argument apart from the words around it. The
// reports pass their own: backticks for a plain one, bold quotes for a
// coloured one.
type ArgFormat func(arg string) string

// Quoted is the plain ArgFormat, and the default when a caller passes none.
func Quoted(arg string) string {
	return fmt.Sprintf("`%s`", arg)
}

func formattedValues(args []any, format ArgFormat) []any {
	if format == nil {
		format = Quoted
	}

	result := make([]any, len(args))
	for i, arg := range args {
		result[i] = format(interfaceToString(arg))
	}
	return result
}

// SetPipedOutput overrides piped-output auto-detection; see colorize.
func SetPipedOutput(val *bool) *bool {
	return colorize.SetPipedOutput(val)
}
