package checker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/oasdiff/oasdiff/checker/rules"
	"github.com/oasdiff/oasdiff/utils"
)

// Level is defined in checker/rules; the aliases keep the checker package's
// public surface unchanged.
type Level = rules.Level

const (
	ERR     = rules.ERR
	WARN    = rules.WARN
	INFO    = rules.INFO
	NONE    = rules.NONE
	INVALID = rules.INVALID
)

func NewLevel(level string) (Level, error) {
	return rules.NewLevel(level)
}

// ProcessSeverityLevels reads a file with severity levels and returns a map of severity levels
func ProcessSeverityLevels(file string) (map[string]Level, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

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
