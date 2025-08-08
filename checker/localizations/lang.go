package localizations

const (
	LangDefault = LangEn
	LangEn      = "en"
	LangRu      = "ru"
	LangPtBr    = "pt-br"
	LangEs      = "es"
)

func GetSupportedLanguages() []string {
	return []string{LangEn, LangRu, LangPtBr, LangEs}
}
