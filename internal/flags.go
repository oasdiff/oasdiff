package internal

import (
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
	"github.com/spf13/viper"
)

type Flags struct {
	v        *viper.Viper
	base     *load.Source
	revision *load.Source
}

func NewFlags() *Flags {
	return &Flags{
		v: viper.New(),
	}
}

func (flags *Flags) toConfig() *diff.Config {
	config := diff.NewConfig().WithExcludeElements(flags.getExcludeElements())
	config.MatchPath = flags.v.GetString("match-path")
	config.UnmatchPath = flags.v.GetString("unmatch-path")
	config.FilterExtension = flags.v.GetString("filter-extension")
	config.PathPrefixBase = flags.v.GetString("prefix-base")
	config.PathPrefixRevision = flags.v.GetString("prefix-revision")
	config.PathStripPrefixBase = flags.v.GetString("strip-prefix-base")
	config.PathStripPrefixRevision = flags.v.GetString("strip-prefix-revision")
	config.IncludePathParams = flags.v.GetBool("include-path-params")

	return config
}

func (flags *Flags) getViper() *viper.Viper {
	return flags.v
}

func (flags *Flags) getAttributes() []string {
	return flags.v.GetStringSlice("attributes")
}

func (flags *Flags) getComposed() bool {
	return flags.v.GetBool("composed")
}

func (flags *Flags) getBase() *load.Source {
	return flags.base
}

func (flags *Flags) getRevision() *load.Source {
	return flags.revision
}

func (flags *Flags) getFlattenAllOf() bool {
	return flags.v.GetBool("flatten-allof") || flags.v.GetBool("flatten")
}

func (flags *Flags) getFlattenParams() bool {
	return flags.v.GetBool("flatten-params")
}

func (flags *Flags) getCaseInsensitiveHeaders() bool {
	return flags.v.GetBool("case-insensitive-headers")
}

func (flags *Flags) getIncludeChecks() []string {
	return fixViperStringSlice(flags.v.GetStringSlice("include-checks"))
}

func (flags *Flags) getDeprecationDaysBeta() uint {
	return flags.v.GetUint("deprecation-days-beta")
}

func (flags *Flags) getDeprecationDaysStable() uint {
	return flags.v.GetUint("deprecation-days-stable")
}

func (flags *Flags) getLang() string {
	return flags.v.GetString("lang")
}

func (flags *Flags) getColor() string {
	return flags.v.GetString("color")
}

func (flags *Flags) getWarnIgnoreFile() string {
	return flags.v.GetString("warn-ignore")
}

func (flags *Flags) getErrIgnoreFile() string {
	return flags.v.GetString("err-ignore")
}

func (flags *Flags) getFormat() string {
	return flags.v.GetString("format")
}

func (flags *Flags) getFailOn() string {
	return flags.v.GetString("fail-on")
}

func (flags *Flags) getLevel() string {
	return flags.v.GetString("level")
}

func (flags *Flags) getFailOnDiff() bool {
	return flags.v.GetBool("fail-on-diff")
}

func (flags *Flags) getSeverityLevelsFile() string {
	return flags.v.GetString("severity-levels")
}

func (flags *Flags) getExcludeElements() []string {
	return fixViperStringSlice(flags.v.GetStringSlice("exclude-elements"))
}

func (flags *Flags) setBase(source *load.Source) {
	flags.base = source
}

func (flags *Flags) setRevision(source *load.Source) {
	flags.revision = source
}

func (flags *Flags) addExcludeElements(element string) {
	flags.v.Set("exclude-elements", append(flags.getExcludeElements(), element))
}

func (flags *Flags) getSeverity() []string {
	return fixViperStringSlice(flags.v.GetStringSlice("severity"))
}

func (flags *Flags) getTags() []string {
	return fixViperStringSlice(flags.v.GetStringSlice("tags"))
}
