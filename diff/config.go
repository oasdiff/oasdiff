package diff

import (
	"github.com/oasdiff/oasdiff/utils"
)

// Config includes various settings to control the diff
type Config struct {
	MatchPath               string
	UnmatchPath             string
	FilterExtension         string
	PathPrefixBase          string
	PathPrefixRevision      string
	PathStripPrefixBase     string
	PathStripPrefixRevision string
	ExcludeElements         utils.StringSet
	ExcludeExtensions       utils.StringSet
	IncludePathParams       bool
}

const (
	ExcludeExamplesOption    = "examples"
	ExcludeDescriptionOption = "description"
	ExcludeEndpointsOption   = "endpoints"
	ExcludeTitleOption       = "title"
	ExcludeSummaryOption     = "summary"
	ExcludeExtensionsOption  = "extensions"
)

func GetExcludeDiffOptions() []string {
	return []string{
		ExcludeExamplesOption,
		ExcludeDescriptionOption,
		ExcludeEndpointsOption,
		ExcludeTitleOption,
		ExcludeSummaryOption,
		ExcludeExtensionsOption,
	}
}

// NewConfig returns a default configuration
func NewConfig() *Config {
	return &Config{
		ExcludeElements:   utils.StringSet{},
		ExcludeExtensions: utils.StringSet{},
	}
}

func (config *Config) WithExcludeElements(excludeElements []string) *Config {
	config.ExcludeElements = utils.StringSetFromSlice(excludeElements)
	return config
}

func (config *Config) WithExcludeExtensions(excludeExtensions []string) *Config {
	config.ExcludeExtensions = utils.StringSetFromSlice(excludeExtensions)
	return config
}

func (config *Config) IsExcludeExamples() bool {
	return config.ExcludeElements.Contains(ExcludeExamplesOption)
}

func (config *Config) IsExcludeDescription() bool {
	return config.ExcludeElements.Contains(ExcludeDescriptionOption)
}

func (config *Config) IsExcludeEndpoints() bool {
	return config.ExcludeElements.Contains(ExcludeEndpointsOption)
}

func (config *Config) IsExcludeTitle() bool {
	return config.ExcludeElements.Contains(ExcludeTitleOption)
}

func (config *Config) IsExcludeSummary() bool {
	return config.ExcludeElements.Contains(ExcludeSummaryOption)
}

func (config *Config) IsExcludeExtensions() bool {
	return config.ExcludeElements.Contains(ExcludeExtensionsOption)
}

// IsExcludedExtension checks if a specific extension name should be excluded from diff
func (config *Config) IsExcludedExtension(name string) bool {
	return config.ExcludeExtensions.Contains(name)
}

const (
	SunsetExtension          = "x-sunset"
	XStabilityLevelExtension = "x-stability-level"
	XExtensibleEnumExtension = "x-extensible-enum"
)
