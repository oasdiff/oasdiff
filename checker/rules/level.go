package rules

import (
	"fmt"

	"github.com/TwiN/go-color"

	"github.com/oasdiff/oasdiff/colorize"
)

type Level int

const (
	ERR     Level = 3
	WARN    Level = 2
	INFO    Level = 1
	NONE    Level = 0
	INVALID Level = -1
)

func NewLevel(level string) (Level, error) {
	switch level {
	case "ERR", "err":
		return ERR, nil
	case "WARN", "warn":
		return WARN, nil
	case "INFO", "info":
		return INFO, nil
	case "NONE", "none":
		return NONE, nil
	}
	return INVALID, fmt.Errorf("invalid level %s", level)
}

func (level Level) StringCond(colorMode colorize.ColorMode) string {
	if colorize.IsColorEnabled(colorMode) {
		return level.PrettyString()
	}
	return level.String()
}

func (level Level) String() string {
	switch level {
	case ERR:
		return "error"
	case WARN:
		return "warning"
	case INFO:
		return "info"
	default:
		return "issue"
	}
}

func (level Level) PrettyString() string {
	levelName := level.String()
	switch level {
	case ERR:
		return color.InRed(levelName)
	case WARN:
		return color.InPurple(levelName)
	case INFO:
		return color.InCyan(levelName)
	default:
		return color.InGray(levelName)
	}
}

func (level Level) IsBreaking() bool {
	return level == ERR || level == WARN
}
