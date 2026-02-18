package checker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TwiN/go-color"
	"github.com/oasdiff/oasdiff/utils"
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

func (level Level) StringCond(colorMode ColorMode) string {
	if isColorEnabled(colorMode) {
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

// ProcessSeverityLevels reads a file with severity levels and returns a map of severity levels
func ProcessSeverityLevels(file string) (map[string]Level, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return GetSeverityLevels(f)
}

// GetSeverityLevels reads severity levels from a reader and returns a map of severity levels
func GetSeverityLevels(source io.Reader) (map[string]Level, error) {

	result := map[string]Level{}

	validIds := utils.StringSetFromSlice(GetAllRuleIds())

	scanner := bufio.NewScanner(source)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		frags := strings.Fields(line)

		if len(frags) != 2 {
			return nil, fmt.Errorf("invalid line #%d: %s", lineNum, line)
		}

		id := frags[0]
		if !validIds.Contains(id) {
			return nil, fmt.Errorf("invalid rule id %q on line %d", id, lineNum)
		}

		level, err := NewLevel(frags[1])
		if err != nil {
			return nil, fmt.Errorf("invalid level %q on line %d", frags[1], lineNum)
		}

		result[id] = level
	}

	return result, nil
}
