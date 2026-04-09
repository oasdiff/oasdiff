package checker

import (
	"fmt"

	"github.com/TwiN/go-color"
)

type ColorMode int

const (
	ColorAlways ColorMode = iota
	ColorNever
	ColorAuto
	ColorInvalid
)

func GetSupportedColorValues() []string {
	return []string{"auto", "always", "never"}
}

func NewColorMode(color string) (ColorMode, error) {
	switch color {
	case "always":
		return ColorAlways, nil
	case "never":
		return ColorNever, nil
	case "auto":
		return ColorAuto, nil
	default:
		return ColorInvalid, fmt.Errorf("invalid color mode: %s", color)
	}
}

func isColorEnabled(colorMode ColorMode) bool {
	switch colorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		return !isPipedOutput()
	default:
		return false
	}
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
