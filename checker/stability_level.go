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
