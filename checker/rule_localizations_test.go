package checker_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/checker/localizations"
)

// Every rule id has a message and a description in every supported locale.
// The localizer is built with no fallback locale, so a missing key returns
// the key itself instead of silently borrowing the English text, which is
// what lets a gap be observed here rather than shipped as a partly-English
// changelog.
func TestRuleLocalizations(t *testing.T) {
	for _, locale := range []string{localizations.LangEn, localizations.LangEs, localizations.LangPtBr, localizations.LangRu} {
		localizer := localizations.New(locale, "")
		for _, rule := range checker.GetAllRules() {
			message := "messages." + rule.Id
			if localizer.Get(message) == message {
				t.Errorf("%s: no message for %s", locale, rule.Id)
			}
			description := "messages." + rule.Id + "-description"
			if localizer.Get(description) == description {
				t.Errorf("%s: no description for %s", locale, rule.Id)
			}
		}
	}
}
