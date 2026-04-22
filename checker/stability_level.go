package checker

// StabilityLevel represents a threshold for filtering operations by their x-stability-level.
// Higher values include more stability levels (e.g., StabilityLevelDraft includes all levels).
type StabilityLevel int

const (
	// StabilityLevelNone means no stability filtering is applied (default behavior).
	StabilityLevelNone StabilityLevel = iota
	// StabilityLevelDraft includes all operations (stable, beta, alpha, and draft).
	StabilityLevelDraft
	// StabilityLevelAlpha includes stable, beta, and alpha operations.
	StabilityLevelAlpha
	// StabilityLevelBeta includes stable and beta operations.
	StabilityLevelBeta
	// StabilityLevelStable includes only stable operations.
	StabilityLevelStable
)

// DefaultStabilityLevel is the default stability level when none is specified.
var DefaultStabilityLevel = StabilityLevelNone

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
		// empty string or unknown → no stability level
		return StabilityLevelNone
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
	if sl == StabilityLevelNone {
		// Default behavior: exclude draft and alpha (original oasdiff behavior)
		return level >= StabilityLevelBeta
	}
	return level >= sl
}
