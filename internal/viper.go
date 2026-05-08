package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/formatters"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"slices"
)

// EnvConfigPath is the environment variable that overrides the default
// config-file lookup. Lower precedence than the --config flag.
const EnvConfigPath = "OASDIFF_CONFIG"

// configRelativePathKeys lists viper keys whose values are file paths.
// When read from a config file, relative values in these keys are
// resolved against the config file's own directory. Absolute values
// and values explicitly set via CLI flag are left alone.
var configRelativePathKeys = []string{
	"err-ignore",
	"warn-ignore",
	"severity-levels",
	"template",
}

type IViper interface {
	SetConfigName(in string)
	SetConfigFile(in string)
	AddConfigPath(in string)
	ReadInConfig() error
	ConfigFileUsed() string
	GetString(key string) string
	Set(key string, value any)
	BindPFlag(key string, flag *pflag.Flag) error
	UnmarshalExact(rawVal any, opts ...viper.DecoderConfigOption) error
}

func RunViper(cmd *cobra.Command, v IViper) *ReturnError {
	if err := readConfFile(cmd, v); err != nil {
		return getErrConfigFileProblem(err)
	}

	if err := validate(v); err != nil {
		return getErrConfigFileProblem(err)
	}

	if err := bindFlags(cmd, v); err != nil {
		return getErrConfigFileProblem(err)
	}

	// After CLI flags are bound, rewrite path-valued config-file values
	// to be relative to the config file's directory (vs. cwd). Skips
	// values where the user explicitly set the corresponding CLI flag —
	// those mean what the user typed.
	resolveConfigRelativePaths(cmd, v)

	return nil
}

// resolveConfigRelativePaths walks the documented path-valued config
// keys and rewrites relative values to be anchored at the config file's
// directory. A config file is self-contained — relative paths in it
// refer to siblings of the config, not arbitrary files at the process's
// cwd.
//
// Skipped per key:
//   - The corresponding CLI flag was explicitly set by the user
//     (Changed=true). Their typed path means what they typed.
//   - The value is empty, or already absolute.
//
// Skipped overall when no config file was loaded.
func resolveConfigRelativePaths(cmd *cobra.Command, v IViper) {
	configFile := v.ConfigFileUsed()
	if configFile == "" {
		return
	}
	configDir := filepath.Dir(configFile)

	for _, key := range configRelativePathKeys {
		if cmd != nil {
			if flag := cmd.Flag(key); flag != nil && flag.Changed {
				continue
			}
		}
		val := v.GetString(key)
		if val == "" || filepath.IsAbs(val) {
			continue
		}
		v.Set(key, filepath.Join(configDir, val))
	}
}

// readConfFile loads the oasdiff configuration file. Resolution order:
//
//  1. --config <path> flag (when set, the file MUST exist; missing or malformed file is an error)
//  2. OASDIFF_CONFIG environment variable (same semantics)
//  3. Default cwd lookup, in order:
//     a. .oasdiff.{json,yaml,yml,toml,hcl} (preferred — dotfile convention)
//     b. oasdiff.{json,yaml,yml,toml,hcl} (legacy — kept for back-compat, still works without warning)
//
// Returns nil when no config is found via the default lookup; only the two
// explicit-override paths surface "file not found" as an error.
func readConfFile(cmd *cobra.Command, v IViper) error {
	if path := explicitConfigPath(cmd); path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("read error: %s", err)
		}
		return nil
	}

	// Default lookup: try .oasdiff.* first (preferred), then oasdiff.*
	// (legacy back-compat) — cwd-scoped, all viper-supported extensions.
	for _, name := range []string{".oasdiff", "oasdiff"} {
		v.SetConfigName(name)
		v.AddConfigPath(".")

		err := v.ReadInConfig()
		if err == nil {
			return nil
		}
		if _, notFound := err.(viper.ConfigFileNotFoundError); notFound {
			continue // try the next candidate
		}
		// File found but malformed — surface the error.
		return fmt.Errorf("read error: %s", err)
	}

	// No config file found in cwd; not an error.
	return nil
}

// explicitConfigPath returns the explicit config path requested by the
// caller via --config or OASDIFF_CONFIG. The flag wins when both are set.
// Returns "" when neither is set (caller falls back to cwd lookup).
func explicitConfigPath(cmd *cobra.Command) string {
	if cmd != nil {
		if flag := cmd.Flag("config"); flag != nil {
			if path := flag.Value.String(); path != "" {
				return path
			}
		}
	}
	return os.Getenv(EnvConfigPath)
}

func bindFlags(cmd *cobra.Command, v IViper) error {
	var result error
	persitentFlags := cmd.PersistentFlags()
	persitentFlags.VisitAll(func(flag *pflag.Flag) {
		name := flag.Name
		if err := v.BindPFlag(name, persitentFlags.Lookup(name)); err != nil {
			result = fmt.Errorf("error binding flag %q to viper: %w", name, err)
			return
		}
	})

	return result
}

// fixViperStringSlice fixes a limitation in viper that doesn't handle custom flags with multiple values
func fixViperStringSlice(viperString []string) []string {
	// viper returns a slice with a single element if the flag was set with a comma-separated list
	if len(viperString) == 1 {
		return strings.Split(viperString[0], ",")
	}

	return viperString
}

type Config struct {
	Attributes             []string `mapstructure:"attributes"`
	Composed               bool     `mapstructure:"composed"`
	FlattenAllof           bool     `mapstructure:"flatten-allof"`
	FlattenParams          bool     `mapstructure:"flatten-params"`
	CaseInsensitiveHeaders bool     `mapstructure:"case-insensitive-headers"`
	DeprecationDaysBeta    uint     `mapstructure:"deprecation-days-beta"`
	DeprecationDaysStable  uint     `mapstructure:"deprecation-days-stable"`
	Lang                   string   `mapstructure:"lang"`
	Color                  string   `mapstructure:"color"`
	WarnIgnore             string   `mapstructure:"warn-ignore"`
	ErrIgnore              string   `mapstructure:"err-ignore"`
	Format                 string   `mapstructure:"format"`
	FailOn                 string   `mapstructure:"fail-on"`
	Level                  string   `mapstructure:"level"`
	FailOnDiff             bool     `mapstructure:"fail-on-diff"`
	SeverityLevels         string   `mapstructure:"severity-levels"`
	ExcludeElements        []string `mapstructure:"exclude-elements"`
	ExcludeExtensions      []string `mapstructure:"exclude-extensions"`
	Severity               []string `mapstructure:"severity"`
	Tags                   []string `mapstructure:"tags"`
	MatchPath              string   `mapstructure:"match-path"`
	UnmatchPath            string   `mapstructure:"unmatch-path"`
	FilterExtension        string   `mapstructure:"filter-extension"`
	PrefixBase             string   `mapstructure:"prefix-base"`
	PrefixRevision         string   `mapstructure:"prefix-revision"`
	StripPrefixBase        string   `mapstructure:"strip-prefix-base"`
	StripPrefixRevision    string   `mapstructure:"strip-prefix-revision"`
	IncludePathParams      bool     `mapstructure:"include-path-params"`
	AllowExternalRefs      bool     `mapstructure:"allow-external-refs"`
	Template               string   `mapstructure:"template"`
}

// validate checks that each of the provided configuration values is one of the generally accepted values
// note that validataion ignores the specific sub-command that was used and is therefor not as strict as the command-specific validation
func validate(v IViper) error {
	var config Config

	if err := v.UnmarshalExact(&config); err != nil {
		return fmt.Errorf("validation error: %s", err)
	}

	if err := validateString(localizations.GetSupportedLanguages(), config.Lang, "lang"); err != nil {
		return err
	}

	if err := validateString(checker.GetSupportedColorValues(), config.Color, "color"); err != nil {
		return err
	}

	if err := validateString(formatters.GetSupportedFormats(), config.Format, "format"); err != nil {
		return err
	}

	if err := validateString(GetSupportedLevels(), config.FailOn, "fail-on"); err != nil {
		return err
	}

	if err := validateString(GetSupportedLevels(), config.Level, "level"); err != nil {
		return err
	}

	if err := validateStrings(diff.GetExcludeDiffOptions(), config.ExcludeElements, "exclude-elements"); err != nil {
		return err
	}

	if err := validateStrings(GetSupportedLevelsLower(), config.Severity, "severity"); err != nil {
		return err
	}

	if err := validateStrings(getAllTags(), config.Tags, "tags"); err != nil {
		return err
	}

	return nil
}

func validateStrings(allowedValues []string, values []string, name string) error {
	for _, value := range values {
		if err := validateString(allowedValues, value, name); err != nil {
			return err
		}
	}
	return nil
}

func validateString(allowedValues []string, value string, name string) error {
	if value == "" {
		return nil
	}

	if slices.Contains(allowedValues, value) {
		return nil
	}

	return fmt.Errorf("invalid %s %q, allowed values: %v", name, value, strings.Join(allowedValues, ", "))
}
