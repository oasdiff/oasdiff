package checker

import "log"

type Config struct {
	Checks              BackwardCompatibilityChecks
	MinSunsetBetaDays   uint
	MinSunsetStableDays uint
	LogLevels           map[string]Level
	Attributes          []string
	StabilityLevel      StabilityLevel
}

const (
	DefaultBetaDeprecationDays   = uint(0)
	DefaultStableDeprecationDays = uint(0)
)

// NewConfig creates a new configuration with default values.
func NewConfig(checks BackwardCompatibilityChecks) *Config {
	return &Config{
		Checks:              checks,
		LogLevels:           GetCheckLevels(),
		MinSunsetBetaDays:   DefaultBetaDeprecationDays,
		MinSunsetStableDays: DefaultStableDeprecationDays,
		StabilityLevel:      DefaultStabilityLevel,
	}
}

// WithOptionalCheck adds a check to the list of optional checks.
func (config *Config) WithOptionalCheck(id string) *Config {
	return config.WithOptionalChecks([]string{id})
}

// WithOptionalChecks overrides the log level of the given checks to ERR so they will appear in `oasdiff breaking`
func (config *Config) WithOptionalChecks(ids []string) *Config {
	for _, id := range ids {
		config.setLogLevel(id, ERR)
	}
	return config
}

func (config *Config) WithSeverityLevels(severityLevels map[string]Level) *Config {
	for id, level := range severityLevels {
		config.setLogLevel(id, level)
	}

	return config
}

// WithDeprecation sets the number of days before sunset for deprecation warnings.
func (config *Config) WithDeprecation(deprecationDaysBeta uint, deprecationDaysStable uint) *Config {
	config.MinSunsetBetaDays = deprecationDaysBeta
	config.MinSunsetStableDays = deprecationDaysStable
	return config
}

// WithSingleCheck sets a single check to be used.
func (config *Config) WithSingleCheck(check BackwardCompatibilityCheck) *Config {
	return config.WithChecks(BackwardCompatibilityChecks{check})
}

// WithChecks sets a list of checks to be used.
func (config *Config) WithChecks(checks BackwardCompatibilityChecks) *Config {
	config.Checks = checks
	return config
}

// WithAttributes sets a list of attributes to be used.
func (config *Config) WithAttributes(attributes []string) *Config {
	config.Attributes = attributes
	return config
}

func (config *Config) getLogLevel(checkId string) Level {
	level, ok := config.LogLevels[checkId]

	if !ok {
		log.Fatal("failed to get log level with invalid check id: ", checkId)
	}

	return level

}

func (config *Config) setLogLevel(checkId string, level Level) {
	if _, ok := config.LogLevels[checkId]; !ok {
		log.Fatal("failed to set log level with invalid check id: ", checkId)
	}

	config.LogLevels[checkId] = level
}
