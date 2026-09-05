package checker

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/oasdiff/oasdiff/checker/localizations"
	"github.com/stretchr/testify/require"
)

var updateBoundMessages = flag.Bool("update-bound-messages", false,
	"rewrite the generated bound-rule entries in localizations_src (run via 'make bound-messages')")

// boundMessageTemplates holds, per locale, the message and description
// patterns for the generated bound rules. {kw} stands for the keyword.
// The shared bound-set comment is a single hand-maintained entry per locale
// and is not generated.
type boundTemplates struct {
	// message templates by scope and action
	message map[[2]string]string
	// description templates by scope and action
	description map[[2]string]string
}

var boundMessageTemplates = map[string]boundTemplates{
	localizations.LangEn: {
		message: map[[2]string]string{
			{"body", "set"}:        "the request's body {kw} was set to %s",
			{"body", "unset"}:      "the request's body {kw} was unset from %s",
			{"property", "set"}:    "the %s request property's {kw} was set to %s",
			{"property", "unset"}:  "the %s request property's {kw} was unset from %s",
			{"parameter", "set"}:   "for the %s request parameter %s, the {kw} was set to %s",
			{"parameter", "unset"}: "for the %s request parameter %s, the {kw} was unset from %s",
			{"rbody", "set"}:       "the response's body {kw} was set to %s",
			{"rbody", "unset"}:     "the response's body {kw} was unset from %s",
			{"rproperty", "set"}:   "the %s response property's {kw} was set to %s for the response status %s",
			{"rproperty", "unset"}: "the %s response property's {kw} was unset from %s for the response status %s",
			{"header", "set"}:      "the %s response header's {kw} was set to %s for the status %s",
			{"header", "unset"}:    "the %s response header's {kw} was unset from %s for the status %s",
		},
		description: map[[2]string]string{
			{"body", ""}:      "request body {kw} {action}",
			{"property", ""}:  "request property {kw} {action}",
			{"parameter", ""}: "request parameter {kw} {action}",
			{"rbody", ""}:     "response body {kw} {action}",
			{"rproperty", ""}: "response property {kw} {action}",
			{"header", ""}:    "response header {kw} {action}",
		},
	},
	localizations.LangEs: {
		message: map[[2]string]string{
			{"body", "set"}:        "el valor {kw} del cuerpo de solicitud fue establecido en %s",
			{"body", "unset"}:      "el valor {kw} del cuerpo de solicitud fue removido de %s",
			{"property", "set"}:    "el valor {kw} de la propiedad de solicitud %s fue establecido en %s",
			{"property", "unset"}:  "el valor {kw} de la propiedad de solicitud %s fue removido de %s",
			{"parameter", "set"}:   "para el parámetro %s de solicitud %s, el valor {kw} fue establecido en %s",
			{"parameter", "unset"}: "para el parámetro %s de solicitud %s, el valor {kw} fue removido de %s",
			{"rbody", "set"}:       "el valor {kw} del cuerpo de respuesta fue establecido en %s",
			{"rbody", "unset"}:     "el valor {kw} del cuerpo de respuesta fue removido de %s",
			{"rproperty", "set"}:   "el valor {kw} de la propiedad de respuesta %s fue establecido en %s para el estado %s",
			{"rproperty", "unset"}: "el valor {kw} de la propiedad de respuesta %s fue removido de %s para el estado %s",
			{"header", "set"}:      "el valor {kw} del encabezado de respuesta %s fue establecido en %s para el estado %s",
			{"header", "unset"}:    "el valor {kw} del encabezado de respuesta %s fue removido de %s para el estado %s",
		},
		description: map[[2]string]string{
			{"body", ""}:      "valor {kw} del cuerpo de solicitud {action}",
			{"property", ""}:  "valor {kw} de la propiedad de solicitud {action}",
			{"parameter", ""}: "valor {kw} del parámetro de solicitud {action}",
			{"rbody", ""}:     "valor {kw} del cuerpo de respuesta {action}",
			{"rproperty", ""}: "valor {kw} de la propiedad de respuesta {action}",
			{"header", ""}:    "valor {kw} del encabezado de respuesta {action}",
		},
	},
	localizations.LangPtBr: {
		message: map[[2]string]string{
			{"body", "set"}:        "o valor {kw} do corpo da requisição foi definido como %s",
			{"body", "unset"}:      "o valor {kw} do corpo da requisição foi removido de %s",
			{"property", "set"}:    "o valor {kw} da propriedade de requisição %s foi definido como %s",
			{"property", "unset"}:  "o valor {kw} da propriedade de requisição %s foi removido de %s",
			{"parameter", "set"}:   "no parâmetro de requisição do tipo %s e nome %s, o valor {kw} foi definido como %s",
			{"parameter", "unset"}: "no parâmetro de requisição do tipo %s e nome %s, o valor {kw} foi removido de %s",
			{"rbody", "set"}:       "o valor {kw} do corpo da resposta foi definido como %s",
			{"rbody", "unset"}:     "o valor {kw} do corpo da resposta foi removido de %s",
			{"rproperty", "set"}:   "o valor {kw} da propriedade de resposta %s foi definido como %s para o status %s",
			{"rproperty", "unset"}: "o valor {kw} da propriedade de resposta %s foi removido de %s para o status %s",
			{"header", "set"}:      "o valor {kw} do cabeçalho de resposta %s foi definido como %s para o status %s",
			{"header", "unset"}:    "o valor {kw} do cabeçalho de resposta %s foi removido de %s para o status %s",
		},
		description: map[[2]string]string{
			{"body", ""}:      "valor {kw} do corpo da requisição {action}",
			{"property", ""}:  "valor {kw} da propriedade de requisição {action}",
			{"parameter", ""}: "valor {kw} do parâmetro da requisição {action}",
			{"rbody", ""}:     "valor {kw} do corpo da resposta {action}",
			{"rproperty", ""}: "valor {kw} da propriedade de resposta {action}",
			{"header", ""}:    "valor {kw} do cabeçalho de resposta {action}",
		},
	},
	localizations.LangRu: {
		message: map[[2]string]string{
			{"body", "set"}:        "у тела запроса задано значение {kw} в %s",
			{"body", "unset"}:      "значение {kw} у тела запроса удалено, предыдущее значение - %s",
			{"property", "set"}:    "у поля запроса %s задано значение {kw} в %s",
			{"property", "unset"}:  "значение {kw} у поля запроса %s удалено, предыдущее значение - %s",
			{"parameter", "set"}:   "в %s параметре запроса %s задано значение {kw} в %s",
			{"parameter", "unset"}: "значение {kw} в %s параметре запроса %s удалено, предыдущее значение - %s",
			{"rbody", "set"}:       "у тела ответа задано значение {kw} в %s",
			{"rbody", "unset"}:     "значение {kw} у тела ответа удалено, предыдущее значение - %s",
			{"rproperty", "set"}:   "у поля ответа %s задано значение {kw} в %s, для ответа со статусом %s",
			{"rproperty", "unset"}: "значение {kw} у поля ответа %s удалено, предыдущее значение - %s, для ответа со статусом %s",
			{"header", "set"}:      "у заголовка ответа %s задано значение {kw} в %s для статуса %s",
			{"header", "unset"}:    "значение {kw} у заголовка ответа %s удалено, предыдущее значение - %s, для статуса %s",
		},
		description: map[[2]string]string{
			{"body", ""}:      "{action} значение {kw} тела запроса",
			{"property", ""}:  "{action} значение {kw} поля запроса",
			{"parameter", ""}: "{action} значение {kw} параметра запроса",
			{"rbody", ""}:     "{action} значение {kw} тела ответа",
			{"rproperty", ""}: "{action} значение {kw} поля ответа",
			{"header", ""}:    "{action} значение {kw} заголовка ответа",
		},
	},
}

// action words for the description templates, per locale
var boundActionWords = map[string]map[string]string{
	localizations.LangEn:   {"set": "set", "unset": "unset"},
	localizations.LangEs:   {"set": "establecido", "unset": "removido"},
	localizations.LangPtBr: {"set": "definido", "unset": "removido"},
	localizations.LangRu:   {"set": "установлено", "unset": "удалено"},
}

// templateScope maps a rule's direction and scope to the template key:
// response body and property phrase differently from their request twins.
func templateScope(direction Direction, scope string) string {
	if direction == DirectionResponse && (scope == "body" || scope == "property") {
		return "r" + scope
	}
	return scope
}

// displayKeyword is the keyword as the messages spell it: max and min keep
// the short forms the hand-written family uses; every other keyword reads as
// its schema field name.
func displayKeyword(keyword string) string {
	switch keyword {
	case "maximum":
		return "max"
	case "minimum":
		return "min"
	}
	return keyword
}

type boundMessage struct {
	key  string
	text string
}

// boundMessages derives the localization entries for every generated bound
// rule in the given locale, from the same tables that generate the rules.
func boundMessages(locale string) []boundMessage {
	templates := boundMessageTemplates[locale]
	words := boundActionWords[locale]

	var entries []boundMessage
	for _, spec := range boundSpecs {
		for _, direction := range []Direction{DirectionRequest, DirectionResponse} {
			for _, scope := range boundScopes(direction) {
				for _, action := range boundActions {
					id := boundRuleId(direction, scope, spec.idName, action.action)
					if handWrittenIds()[id] {
						continue
					}
					ts := templateScope(direction, scope)
					kw := displayKeyword(spec.keyword)
					message := strings.ReplaceAll(templates.message[[2]string{ts, action.action}], "{kw}", kw)
					description := templates.description[[2]string{ts, ""}]
					description = strings.ReplaceAll(description, "{kw}", kw)
					description = strings.ReplaceAll(description, "{action}", words[action.action])
					entries = append(entries, boundMessage{id, message}, boundMessage{id + "-description", description})
				}
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	return entries
}

const boundMessagesHeader = "# generated bound-rule messages: edit the templates in checker/bound_messages_test.go and run 'make bound-messages'; do not edit these entries by hand"

// The localized message and description of every generated bound rule equal
// what the templates derive, in every locale, checked through the localizer.
// A new table row or template change fails here until 'make bound-messages'
// regenerates the entries.
func TestBoundMessages(t *testing.T) {
	if *updateBoundMessages {
		writeBoundMessages(t)
		t.Skip("regenerated; run 'make localize' and re-run the test")
	}
	for _, locale := range []string{localizations.LangEn, localizations.LangEs, localizations.LangPtBr, localizations.LangRu} {
		localizer := localizations.New(locale, "")
		for _, entry := range boundMessages(locale) {
			require.Equal(t, entry.text, localizer.Get("messages."+entry.key),
				"%s/%s differs from its template; run 'make bound-messages'", locale, entry.key)
		}
	}
}

// writeBoundMessages rewrites the generated entries in each locale's
// messages.yaml: lines carrying a generated key (and the header comment) are
// dropped wherever they are, and the freshly derived block is appended.
func writeBoundMessages(t *testing.T) {
	t.Helper()
	for _, locale := range []string{localizations.LangEn, localizations.LangEs, localizations.LangPtBr, localizations.LangRu} {
		entries := boundMessages(locale)
		generated := map[string]bool{}
		for _, entry := range entries {
			generated[entry.key] = true
		}

		path := filepath.Join("localizations_src", locale, "messages.yaml")
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		var kept []string
		for line := range strings.SplitSeq(strings.TrimRight(string(raw), "\n"), "\n") {
			key, _, found := strings.Cut(line, ": ")
			if line == boundMessagesHeader || (found && generated[key]) {
				continue
			}
			kept = append(kept, line)
		}
		for slices.Contains([]string{""}, kept[len(kept)-1]) {
			kept = kept[:len(kept)-1]
		}

		kept = append(kept, "", boundMessagesHeader)
		for _, entry := range entries {
			kept = append(kept, fmt.Sprintf("%s: %s", entry.key, entry.text))
		}
		require.NoError(t, os.WriteFile(path, []byte(strings.Join(kept, "\n")+"\n"), 0o644))
	}
}
