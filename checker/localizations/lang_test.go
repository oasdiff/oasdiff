package localizations_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/stretchr/testify/require"
)

func TestLang_Exists(t *testing.T) {
	require.Equal(t, []string{localizations.LangEn, localizations.LangRu, localizations.LangPtBr}, localizations.GetSupportedLanguages())
}
