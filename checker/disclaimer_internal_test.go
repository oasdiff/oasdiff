package checker

import (
	"regexp"
	"testing"

	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/stretchr/testify/require"
)

func allDisclaimers(t *testing.T) []Disclaimer {
	t.Helper()

	unnamed := Disclaimer(0).String()
	var all []Disclaimer
	for d := Disclaimer(1); d.String() != unnamed; d++ {
		all = append(all, d)
	}
	require.NotEmpty(t, all)
	return all
}

// Names are a public surface: they appear in output and are the key a caller
// would use to configure a disclaimer. They have to be unique and stable.
func TestDisclaimerNames(t *testing.T) {
	kebab := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	seen := map[string]bool{}
	for _, d := range allDisclaimers(t) {
		name := d.String()
		require.Truef(t, kebab.MatchString(name), "disclaimer name %q is not kebab-case", name)
		require.Falsef(t, seen[name], "duplicate disclaimer name %q", name)
		seen[name] = true
	}
}

// Every disclaimer explains itself, in every language, or it reaches a user as
// a bare identifier.
func TestDisclaimersAreLocalized(t *testing.T) {
	for _, lang := range localizations.GetSupportedLanguages() {
		localizer := NewLocalizer(lang)
		for _, d := range allDisclaimers(t) {
			key := commentId(d.String())
			require.NotEqualf(t, key, localizer(key), "disclaimer %q has no %s text", d, lang)
		}
	}
}
