package checker

import "log"

type Config struct {
	Checks              BackwardCompatibilityChecks
	MinSunsetBetaDays   uint
	MinSunsetStableDays uint
	LogLevels           map[string]Level
	// overriddenLevels holds the ids whose level the caller set, so a
	// disclaimer knows not to lower a level that was chosen deliberately.
	// Distinct from comparing against the registry, which cannot tell a rule
	// set to the level it already had from one never set at all.
	overriddenLevels map[string]bool
	Attributes       []string
	StabilityLevel   StabilityLevel
}

const (
	DefaultBetaDeprecationDays   = uint(0)
	DefaultStableDeprecationDays = uint(0)
)

// Option configures a Config during NewConfig. Options compose: each
// receives the Config after defaults and prior options have been
// applied. Use the With* constructors below to obtain Options.
type Option func(*Config)

// NewConfig creates a new configuration with default values, then
// applies the given options in order.
func NewConfig(checks BackwardCompatibilityChecks, opts ...Option) *Config {
	c := &Config{
		Checks:              checks,
		LogLevels:           GetCheckLevels(),
		MinSunsetBetaDays:   DefaultBetaDeprecationDays,
		MinSunsetStableDays: DefaultStableDeprecationDays,
		StabilityLevel:      DefaultStabilityLevel,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithSeverityLevels overrides per-check log levels from the given map.
func WithSeverityLevels(severityLevels map[string]Level) Option {
	return func(c *Config) {
		for id, level := range severityLevels {
			c.setLogLevel(id, level)
		}
	}
}

// WithDeprecation sets the number of days before sunset for deprecation warnings.
func WithDeprecation(deprecationDaysBeta uint, deprecationDaysStable uint) Option {
	return func(c *Config) {
		c.MinSunsetBetaDays = deprecationDaysBeta
		c.MinSunsetStableDays = deprecationDaysStable
	}
}

// WithSingleCheck sets a single check to be used (replaces the
// constructor's checks argument).
func WithSingleCheck(check BackwardCompatibilityCheck) Option {
	return WithChecks(BackwardCompatibilityChecks{check})
}

// WithChecks sets the list of checks to be used (replaces the
// constructor's checks argument).
func WithChecks(checks BackwardCompatibilityChecks) Option {
	return func(c *Config) {
		c.Checks = checks
	}
}

// WithAttributes sets the list of attributes to be used.
func WithAttributes(attributes []string) Option {
	return func(c *Config) {
		c.Attributes = attributes
	}
}

// WithStabilityLevel sets the minimum stability level from a string.
// If the level is empty, the config is unchanged.
func WithStabilityLevel(level string) Option {
	return func(c *Config) {
		if level != "" {
			c.StabilityLevel = ParseStabilityLevel(level)
		}
	}
}

func (config *Config) getLogLevel(checkId string) Level {
	level, ok := config.LogLevels[checkId]

	if !ok {
		log.Fatal("failed to get log level with invalid check id: ", checkId)
	}

	return level

}

func (config *Config) setLogLevel(checkId string, level Level) {
	if config.overriddenLevels == nil {
		config.overriddenLevels = map[string]bool{}
	}
	config.overriddenLevels[checkId] = true
	if _, ok := config.LogLevels[checkId]; !ok {
		log.Fatal("failed to set log level with invalid check id: ", checkId)
	}

	config.LogLevels[checkId] = level
}
