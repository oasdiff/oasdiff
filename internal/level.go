package internal

const (
	LevelErr  = "ERR"
	LevelWarn = "WARN"
	LevelInfo = "INFO"
)

// FailOnMissingVersionBump is the --fail-on value that asks for the versioning
// policy instead of a severity threshold. See docs/VERSIONING.md.
const FailOnMissingVersionBump = "MISSING-VERSION-BUMP"

func GetSupportedLevels() []string {
	return []string{LevelErr, LevelWarn, LevelInfo}
}

func GetBreakingLevels() []string {
	return []string{LevelErr, LevelWarn}
}

// GetFailOnValues returns the --fail-on values of the changelog command, and
// GetBreakingFailOnValues those of the commands that report breaking changes
// only. Each is the command's severity levels plus the versioning policy.
func GetFailOnValues() []string {
	return append(GetSupportedLevels(), FailOnMissingVersionBump)
}

func GetBreakingFailOnValues() []string {
	return append(GetBreakingLevels(), FailOnMissingVersionBump)
}

func GetSupportedLevelsLower() []string {
	return []string{"error", "warn", "info"}
}
