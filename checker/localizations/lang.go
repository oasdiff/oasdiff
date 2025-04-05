package localizations

const (
	LangDefault = LangEn
	LangEn      = "en"
	LangRu      = "ru"
	LangPtBr    = "pt-br"
)

func GetSupportedLanguages() []string {
	return []string{LangEn, LangRu, LangPtBr}
}
