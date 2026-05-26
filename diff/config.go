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
	MatchInlineRefs         bool
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

// Option configures a Config during NewConfig. Options compose: each
// receives the Config after defaults and prior options have been
// applied.
type Option func(*Config)

// NewConfig returns a default configuration, then applies the given
// options in order.
func NewConfig(opts ...Option) *Config {
	c := &Config{
		ExcludeElements:   utils.StringSet{},
		ExcludeExtensions: utils.StringSet{},
		MatchInlineRefs:   true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithMatchInlineRefs controls whether validation-equivalent inline/$ref
// subschemas under anyOf/oneOf are matched as the same branch.
// Default true. Set to false to restore the previous behaviour where an
// inline-to-$ref refactor of an equivalent component is reported as one
// branch added and one branch removed.
func WithMatchInlineRefs(matchInlineRefs bool) Option {
	return func(c *Config) {
		c.MatchInlineRefs = matchInlineRefs
	}
}

// WithExcludeElements sets the elements (description, summary,
// endpoints, examples, extensions, title) to omit from the diff.
func WithExcludeElements(excludeElements []string) Option {
	return func(c *Config) {
		c.ExcludeElements = utils.StringSetFromSlice(excludeElements)
	}
}

// WithExcludeExtensions sets specific OpenAPI extension names to omit
// from the diff (only takes effect when "extensions" is also in
// ExcludeElements).
func WithExcludeExtensions(excludeExtensions []string) Option {
	return func(c *Config) {
		c.ExcludeExtensions = utils.StringSetFromSlice(excludeExtensions)
	}
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
