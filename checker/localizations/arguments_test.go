package localizations_test

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/stretchr/testify/require"
)

var formatVerb = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z]`)

// Every locale's message takes the arguments the English one takes. The checker
// passes one set of arguments whatever the locale, so a message with fewer verbs
// renders %!(EXTRA ...) after the text and one with more renders %!s(MISSING).
func TestLocalizations_EveryLocaleTakesTheSameArguments(t *testing.T) {
	all := localizations.New(localizations.LangDefault, localizations.LangDefault).Localizations

	// the key is "<locale>.messages.<id>"
	patterns := map[string]map[string]string{}
	for key, pattern := range all {
		locale, id, found := strings.Cut(key, ".")
		require.True(t, found, "key %q has no locale", key)
		if patterns[id] == nil {
			patterns[id] = map[string]string{}
		}
		patterns[id][locale] = pattern
	}

	ids := make([]string, 0, len(patterns))
	for id := range patterns {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	mismatched := []string{}
	for _, id := range ids {
		english, ok := patterns[id][localizations.LangDefault]
		if !ok {
			continue
		}
		// the argument's type follows the verb English uses, so that a locale
		// which prints a count with %s is reported like any other mismatch
		verbs := formatVerb.FindAllString(english, -1)
		args := make([]any, len(verbs))
		for i, verb := range verbs {
			if strings.ContainsAny(verb[len(verb)-1:], "bcdoOqUxX") {
				args[i] = 1
			} else {
				args[i] = "x"
			}
		}
		locales := make([]string, 0, len(patterns[id]))
		for locale := range patterns[id] {
			locales = append(locales, locale)
		}
		sort.Strings(locales)
		for _, locale := range locales {
			if text := fmt.Sprintf(patterns[id][locale], args...); strings.Contains(text, "%!") {
				mismatched = append(mismatched, fmt.Sprintf("%s in %s: %s", id, locale, text))
			}
		}
	}
	require.Empty(t, mismatched, "these messages take different arguments than %s", localizations.LangDefault)
}
