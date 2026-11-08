package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/oasdiff/oasdiff/checker/internal/extension"
	"github.com/oasdiff/oasdiff/diff"
)

// StabilityLevel represents a threshold for filtering operations by their x-stability-level.
// Higher values include more stability levels (e.g., StabilityLevelDraft includes all levels).
// The enum is a pure ordering: level >= threshold means "included."
type StabilityLevel int

const (
	// StabilityLevelDraft includes all operations (stable, beta, alpha, and draft).
	StabilityLevelDraft StabilityLevel = iota + 1
	// StabilityLevelAlpha includes stable, beta, and alpha operations.
	StabilityLevelAlpha
	// StabilityLevelBeta includes stable and beta operations.
	StabilityLevelBeta
	// StabilityLevelStable includes only stable operations.
	StabilityLevelStable
)

// DefaultStabilityLevel is the default stability level when none is specified.
// This matches the documented CLI default of "beta".
const DefaultStabilityLevel = StabilityLevelBeta

// ParseStabilityLevel converts a stability string to a StabilityLevel.
// An empty string is treated as stable (implicit stable).
func ParseStabilityLevel(s string) StabilityLevel {
	switch s {
	case StabilityDraft:
		return StabilityLevelDraft
	case StabilityAlpha:
		return StabilityLevelAlpha
	case StabilityBeta:
		return StabilityLevelBeta
	case StabilityStable:
		return StabilityLevelStable
	default:
		// empty string or unknown → treat as stable
		return StabilityLevelStable
	}
}

// IsIncluded returns true if the given stability label meets the configured threshold.
// For example, if the threshold is StabilityLevelDraft, all levels are included.
// If the threshold is StabilityLevelAlpha, draft is excluded but alpha/beta/stable are included.
// An empty stability string is treated as "stable".
func (sl StabilityLevel) IsIncluded(stability string) bool {
	return ParseStabilityLevel(stability) >= sl
}

// GetSupportedStabilityLevels returns the list of valid stability level strings.
func GetSupportedStabilityLevels() []string {
	return []string{StabilityDraft, StabilityAlpha, StabilityBeta, StabilityStable}
}

const (
	StabilityDraft  = "draft"
	StabilityAlpha  = "alpha"
	StabilityBeta   = "beta"
	StabilityStable = "stable"
)

const (
	// Deprecated: use StabilityDraft.
	STABILITY_DRAFT = StabilityDraft
	// Deprecated: use StabilityAlpha.
	STABILITY_ALPHA = StabilityAlpha
	// Deprecated: use StabilityBeta.
	STABILITY_BETA = StabilityBeta
	// Deprecated: use StabilityStable.
	STABILITY_STABLE = StabilityStable
)

func getStabilityLevel(i map[string]any) (string, error) {
	if i == nil || i[diff.XStabilityLevelExtension] == nil {
		return "", nil
	}
	value := i[diff.XStabilityLevelExtension]
	if jsonStabilityRaw, ok := value.(json.RawMessage); ok {
		if err := json.Unmarshal(jsonStabilityRaw, &value); err != nil {
			// bytes that are not valid json are not a level either, and the value
			// is what the message has to report
			return "", &invalidStabilityLevel{Value: strconv.Quote(string(jsonStabilityRaw))}
		}
	}

	stabilityLevel, ok := value.(string)
	if !ok {
		return "", &invalidStabilityLevel{Value: extension.ValueText(value)}
	}

	if stabilityLevel != StabilityDraft &&
		stabilityLevel != StabilityAlpha &&
		stabilityLevel != StabilityBeta &&
		stabilityLevel != StabilityStable {
		return "", &invalidStabilityLevel{Value: strconv.Quote(stabilityLevel)}
	}

	return stabilityLevel, nil
}

func normalizedStability(label string) string {
	if label == "" {
		return StabilityStable
	}
	return label
}

// invalidStabilityLevel reports an x-stability-level value that is not one of
// the levels oasdiff knows.
type invalidStabilityLevel struct {
	// Value is the value the message reports, quoted when it is a string, so
	// that an empty or space-padded level is visible in every output format.
	Value string
}

func (e *invalidStabilityLevel) Error() string {
	return fmt.Sprintf("stability level %s is not one of %s, %s, %s or %s",
		e.Value, StabilityDraft, StabilityAlpha, StabilityBeta, StabilityStable)
}

// stabilityValue is the message argument for an error getStabilityLevel
// returned: the value it refused, so that the localized message holds the
// sentence and the argument holds only the value.
func stabilityValue(err error) string {
	if invalid, ok := errors.AsType[*invalidStabilityLevel](err); ok {
		return invalid.Value
	}
	return err.Error()
}

func getDeprecationDays(config *Config, stability string) uint {
	switch stability {
	case StabilityDraft:
		return 0
	case StabilityAlpha:
		return 0
	case StabilityBeta:
		return config.MinSunsetBetaDays
	case StabilityStable:
		return config.MinSunsetStableDays
	default:
		return config.MinSunsetStableDays
	}
}
