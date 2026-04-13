package checker

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
var DefaultStabilityLevel = StabilityLevelBeta

// ParseStabilityLevel converts a stability string to a StabilityLevel.
// An empty string is treated as stable (implicit stable).
func ParseStabilityLevel(s string) StabilityLevel {
	switch s {
	case STABILITY_DRAFT:
		return StabilityLevelDraft
	case STABILITY_ALPHA:
		return StabilityLevelAlpha
	case STABILITY_BETA:
		return StabilityLevelBeta
	case STABILITY_STABLE:
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
	// Empty string means no x-stability-level → treat as implicit stable, always included
	if stability == "" {
		return true
	}
	level := ParseStabilityLevel(stability)
	return level >= sl
}

// GetSupportedStabilityLevels returns the list of valid stability level strings.
func GetSupportedStabilityLevels() []string {
	return []string{STABILITY_DRAFT, STABILITY_ALPHA, STABILITY_BETA, STABILITY_STABLE}
}
