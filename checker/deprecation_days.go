package checker

// deprecationDays is the sunset grace period for an operation. ok is false
// where its stability level doesn't parse.
func deprecationDays(config *Config, extensions map[string]any) (days uint, ok bool) {
	stability, err := getStabilityLevel(extensions)
	if err != nil {
		return 0, false
	}
	return getDeprecationDays(config, stability), true
}
