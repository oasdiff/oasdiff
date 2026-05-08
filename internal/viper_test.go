package internal_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type ViperMock struct {
	viper.Viper
	ReadInConfigMock func() error
	BindPFlagMock    func(key string, flag *pflag.Flag) error
}

func NewViperMock() *ViperMock {
	result := ViperMock{
		Viper: *viper.GetViper(),
	}
	result.ReadInConfigMock = result.Viper.ReadInConfig
	result.BindPFlagMock = result.Viper.BindPFlag
	return &result
}

func (v *ViperMock) ReadInConfig() error {
	return v.ReadInConfigMock()
}

func (v *ViperMock) BindPFlag(key string, flag *pflag.Flag) error {
	return v.BindPFlagMock(key, flag)
}

func TestViper_ReadInConfigErr(t *testing.T) {
	v := NewViperMock()
	v.ReadInConfigMock = func() error { return errors.New("error") }

	cmd := cobra.Command{}
	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: read error: error")
}

func TestViper_BindPFlagErr(t *testing.T) {
	v := NewViperMock()
	v.BindPFlagMock = func(key string, flag *pflag.Flag) error {
		return errors.New("error")
	}

	cmd := cobra.Command{}
	cmd.PersistentFlags().BoolP("composed", "c", false, "work in 'composed' mode, compare paths in all specs matching base and revision globs")

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: error binding flag \"composed\" to viper: error")
}

func TestViper_InvalidLang(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("lang: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}

func TestViper_InvalidColor(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("color: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid color \"invalid\", allowed values: auto, always, never")
}

func TestViper_InvalidFormat(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("format: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid format \"invalid\", allowed values: yaml, json, text, markup, markdown, singleline, html, githubactions, junit, sarif")
}

func TestViper_InvalidFailOn(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("fail-on: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid fail-on \"invalid\", allowed values: ERR, WARN, INFO")
}

func TestViper_InvalidLevel(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("level: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid level \"invalid\", allowed values: ERR, WARN, INFO")
}

func TestViper_InvalidExcludeElements(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("exclude-elements: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid exclude-elements \"invalid\", allowed values: examples, description, endpoints, title, summary, extensions")
}

func TestViper_InvalidSeverity(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("severity: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid severity \"invalid\", allowed values: error, warn, info")
}

func TestViper_InvalidTags(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("tags: invalid")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid tags \"invalid\", allowed values: request, response, add, remove, change, generalize, specialize, increase, decrease, set, body, parameters, properties, headers, security, components")
}

func TestViper_ValidTags(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("tags: request")))

	cmd := cobra.Command{}

	require.Nil(t, internal.RunViper(&cmd, v))
}

func TestViper_InvalidFlag(t *testing.T) {
	v := NewViperMock()
	v.SetConfigFile("config.yaml")
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader("invalid: value")))

	cmd := cobra.Command{}

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: validation error: decoding failed due to the following error(s):\n\n'internal.Config' has invalid keys: invalid")
}

// ---------------------------------------------------------------------
// Config-file lookup: .oasdiff.* in cwd; --config flag; OASDIFF_CONFIG
// env var; precedence flag > env > default.
//
// Each test uses t.Chdir(t.TempDir()) so it runs in an isolated cwd
// without polluting the test working directory or interfering with
// sibling tests.
// ---------------------------------------------------------------------

// writeFile writes content to path, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

// chdirIsolated moves the test into a fresh temp dir for the duration
// of the test. Returns the temp dir path so callers can construct
// absolute paths to fixtures they create inside it.
func chdirIsolated(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	return dir
}

// TestViper_DefaultLookup_DotOasdiffYaml: when .oasdiff.yaml exists in
// cwd and no override is set, it is loaded. Asserted via a value that
// validate() rejects — error message proves the file was read.
func TestViper_DefaultLookup_DotOasdiffYaml(t *testing.T) {
	chdirIsolated(t)
	writeFile(t, ".oasdiff.yaml", "lang: invalid\n")

	cmd := cobra.Command{}
	require.EqualError(t,
		internal.RunViper(&cmd, viper.New()),
		"failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}

// TestViper_DefaultLookup_LegacyOasdiffYamlStillWorks: the legacy
// oasdiff.{yaml,...} filename remains supported as a back-compat
// fallback for CLI users who had it set up before the .oasdiff.* move.
// .oasdiff.* is preferred, but oasdiff.* keeps working when no dotfile
// is present.
func TestViper_DefaultLookup_LegacyOasdiffYamlStillWorks(t *testing.T) {
	chdirIsolated(t)
	writeFile(t, "oasdiff.yaml", "lang: invalid\n")

	cmd := cobra.Command{}
	require.EqualError(t,
		internal.RunViper(&cmd, viper.New()),
		"failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}

// TestViper_DefaultLookup_DotPreferredOverLegacy: when both .oasdiff.yaml
// and oasdiff.yaml exist in cwd, the dotfile wins. Proves the lookup
// order — preferred first, legacy as fallback only.
func TestViper_DefaultLookup_DotPreferredOverLegacy(t *testing.T) {
	chdirIsolated(t)
	// Dotfile is clean; legacy is not. If the legacy file were loaded,
	// validate() would error.
	writeFile(t, ".oasdiff.yaml", "lang: en\n")
	writeFile(t, "oasdiff.yaml", "lang: invalid\n")

	cmd := cobra.Command{}
	require.Nil(t, internal.RunViper(&cmd, viper.New()))
}

// TestViper_DefaultLookup_NoConfigIsNotAnError: with no .oasdiff.*
// in cwd and no override, RunViper succeeds.
func TestViper_DefaultLookup_NoConfigIsNotAnError(t *testing.T) {
	chdirIsolated(t)

	cmd := cobra.Command{}
	require.Nil(t, internal.RunViper(&cmd, viper.New()))
}

// TestViper_ConfigFlag_ExplicitPath: --config <path> loads the file
// at the given path, regardless of cwd's .oasdiff.*.
func TestViper_ConfigFlag_ExplicitPath(t *testing.T) {
	dir := chdirIsolated(t)
	customPath := filepath.Join(dir, "custom-config.yaml")
	writeFile(t, customPath, "lang: invalid\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", customPath))

	require.EqualError(t,
		internal.RunViper(&cmd, viper.New()),
		"failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}

// TestViper_ConfigFlag_MissingFileIsError: --config pointing at a
// non-existent file is an explicit error (unlike the silent-skip
// behavior of the default lookup).
func TestViper_ConfigFlag_MissingFileIsError(t *testing.T) {
	chdirIsolated(t)

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", "does-not-exist.yaml"))

	err := internal.RunViper(&cmd, viper.New())
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "read error")
}

// TestViper_ConfigFlag_OverridesDefaultLookup: when --config is set,
// the default cwd lookup is skipped — even if .oasdiff.yaml is also
// present at cwd. Proves the flag's path is used, not the default.
func TestViper_ConfigFlag_OverridesDefaultLookup(t *testing.T) {
	dir := chdirIsolated(t)
	// Default-lookup file with a value that WOULD fail validation if loaded.
	writeFile(t, ".oasdiff.yaml", "lang: invalid\n")
	// Explicit-path file that's clean.
	customPath := filepath.Join(dir, "clean.yaml")
	writeFile(t, customPath, "lang: en\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", customPath))

	// No error: the explicit clean.yaml was loaded, not the bad .oasdiff.yaml.
	require.Nil(t, internal.RunViper(&cmd, viper.New()))
}

// TestViper_ConfigEnvVar: OASDIFF_CONFIG points at a config file in
// the absence of --config.
func TestViper_ConfigEnvVar(t *testing.T) {
	dir := chdirIsolated(t)
	customPath := filepath.Join(dir, "env-config.yaml")
	writeFile(t, customPath, "lang: invalid\n")

	t.Setenv("OASDIFF_CONFIG", customPath)

	cmd := cobra.Command{}
	require.EqualError(t,
		internal.RunViper(&cmd, viper.New()),
		"failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}

// TestViper_ConfigFlagWinsOverEnv: when both --config and
// OASDIFF_CONFIG are set, the flag takes precedence.
func TestViper_ConfigFlagWinsOverEnv(t *testing.T) {
	dir := chdirIsolated(t)
	// Env points at a clean file; flag points at a bad one. If env
	// were honored, no error. If flag wins, validation fails.
	envPath := filepath.Join(dir, "env-clean.yaml")
	writeFile(t, envPath, "lang: en\n")
	flagPath := filepath.Join(dir, "flag-bad.yaml")
	writeFile(t, flagPath, "lang: invalid\n")

	t.Setenv("OASDIFF_CONFIG", envPath)

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", flagPath))

	require.EqualError(t,
		internal.RunViper(&cmd, viper.New()),
		"failed to load config file: invalid lang \"invalid\", allowed values: en, ru, pt-br, es")
}
