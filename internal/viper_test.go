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

	require.EqualError(t, internal.RunViper(&cmd, v), "failed to load config file: invalid tags \"invalid\", allowed values: request, response, add, remove, change, generalize, specialize, increase, decrease, set, schema, parameters, requestBody, responses, paths, headers, security, tags, components, existence, requiredness, mutability, type, constraints, values, structure, lifecycle")
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

// ---------------------------------------------------------------------
// Relative-path resolution: path-valued config keys (err-ignore,
// warn-ignore, severity-levels, template) should resolve against the
// config file's directory, not the process's cwd. Matches behaviour
// of ESLint, Prettier, golangci-lint, etc.
// ---------------------------------------------------------------------

// TestViper_RelativePath_ResolvesAgainstConfigDir: when the config
// file lives in a subdirectory and references a sibling file via a
// relative path, the path is rewritten to be anchored at the config's
// directory.
func TestViper_RelativePath_ResolvesAgainstConfigDir(t *testing.T) {
	root := chdirIsolated(t)
	configDir := filepath.Join(root, "subdir")
	require.NoError(t, os.Mkdir(configDir, 0700))

	configPath := filepath.Join(configDir, "cfg.yaml")
	writeFile(t, configPath, "err-ignore: ignore-rules.txt\n")
	// Sibling file next to the config — what the relative path should
	// resolve to.
	ignorePath := filepath.Join(configDir, "ignore-rules.txt")
	writeFile(t, ignorePath, "")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	cmd.PersistentFlags().String("err-ignore", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", configPath))

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))
	require.Equal(t, ignorePath, v.GetString("err-ignore"))
}

// TestViper_RelativePath_AbsoluteValueUnchanged: an absolute path in
// the config file is left alone, regardless of where the config lives.
func TestViper_RelativePath_AbsoluteValueUnchanged(t *testing.T) {
	root := chdirIsolated(t)
	configDir := filepath.Join(root, "subdir")
	require.NoError(t, os.Mkdir(configDir, 0700))

	absoluteIgnore := filepath.Join(root, "elsewhere.txt")
	writeFile(t, absoluteIgnore, "")

	configPath := filepath.Join(configDir, "cfg.yaml")
	writeFile(t, configPath, "err-ignore: "+absoluteIgnore+"\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	cmd.PersistentFlags().String("err-ignore", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", configPath))

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))
	require.Equal(t, absoluteIgnore, v.GetString("err-ignore"))
}

// TestViper_RelativePath_CLIFlagOverridesConfig: when the CLI flag
// is explicitly set by the user, its value wins over the config and
// is NOT rewritten — the user's path means what they typed.
func TestViper_RelativePath_CLIFlagOverridesConfig(t *testing.T) {
	root := chdirIsolated(t)
	configDir := filepath.Join(root, "subdir")
	require.NoError(t, os.Mkdir(configDir, 0700))

	configPath := filepath.Join(configDir, "cfg.yaml")
	writeFile(t, configPath, "err-ignore: from-config.txt\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	cmd.PersistentFlags().String("err-ignore", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("config", configPath))
	// User explicitly passes --err-ignore; this should win and stay
	// as typed (not rewritten relative to configDir).
	require.NoError(t, cmd.PersistentFlags().Set("err-ignore", "from-cli.txt"))

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))
	require.Equal(t, "from-cli.txt", v.GetString("err-ignore"))
}

// TestViper_RelativePath_AllPathKeysHandled: each documented path-valued
// config key gets the same rewrite treatment.
func TestViper_RelativePath_AllPathKeysHandled(t *testing.T) {
	root := chdirIsolated(t)
	configDir := filepath.Join(root, "subdir")
	require.NoError(t, os.Mkdir(configDir, 0700))

	// One YAML setting all four path-valued keys with relative values.
	configPath := filepath.Join(configDir, "cfg.yaml")
	writeFile(t, configPath, `
err-ignore: err.txt
warn-ignore: warn.txt
severity-levels: sev.txt
template: tmpl.tmpl
`)

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("config", "", "")
	for _, k := range []string{"err-ignore", "warn-ignore", "severity-levels", "template"} {
		cmd.PersistentFlags().String(k, "", "")
	}
	require.NoError(t, cmd.PersistentFlags().Set("config", configPath))

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))

	for k, expected := range map[string]string{
		"err-ignore":      filepath.Join(configDir, "err.txt"),
		"warn-ignore":     filepath.Join(configDir, "warn.txt"),
		"severity-levels": filepath.Join(configDir, "sev.txt"),
		"template":        filepath.Join(configDir, "tmpl.tmpl"),
	} {
		require.Equal(t, expected, v.GetString(k), "key %q", k)
	}
}

// TestViper_RelativePath_DefaultLookup_ConfigInCwd: when the config
// file is loaded from cwd via the default lookup (.oasdiff.yaml),
// relative paths still work — they get rewritten to absolute paths
// anchored at cwd, which is functionally equivalent to the
// pre-existing cwd-relative behaviour. No behaviour change for the
// default case.
func TestViper_RelativePath_DefaultLookup_ConfigInCwd(t *testing.T) {
	root := chdirIsolated(t)
	writeFile(t, ".oasdiff.yaml", "err-ignore: ignore-rules.txt\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("err-ignore", "", "")

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))
	// Path is rewritten to an absolute path under cwd. Functionally
	// equivalent to "ignore-rules.txt" relative to cwd.
	require.Equal(t, filepath.Join(root, "ignore-rules.txt"), v.GetString("err-ignore"))
}

// TestViper_RelativePath_BackCompat_LegacyOasdiffYamlInCwd: explicitly
// covers the population this fix could conceivably regress —
// long-standing CLI users with the legacy `oasdiff.yaml` filename in
// cwd referencing a sibling file via a relative path. They were
// reading <cwd>/<file>; after the fix, the path is rewritten to
// <abs cwd>/<file>. Same file, same behaviour.
func TestViper_RelativePath_BackCompat_LegacyOasdiffYamlInCwd(t *testing.T) {
	root := chdirIsolated(t)
	writeFile(t, "oasdiff.yaml", "err-ignore: my-rules.txt\n")

	cmd := cobra.Command{}
	cmd.PersistentFlags().String("err-ignore", "", "")

	v := viper.New()
	require.Nil(t, internal.RunViper(&cmd, v))
	require.Equal(t, filepath.Join(root, "my-rules.txt"), v.GetString("err-ignore"))
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
