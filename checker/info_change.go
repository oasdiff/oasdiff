package checker

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
)

// InfoChange represents a change in the Info Section: https://swagger.io/specification/#info-object
// It carries no path or operation: the subject is the document as a whole.
type InfoChange struct {
	CommonChange

	Id      string
	Args    []any
	Comment string
	Level   Level
}

// WithSources returns a copy of the InfoChange with BaseSource and RevisionSource populated
func (c InfoChange) WithSources(baseSource, revisionSource *Source) InfoChange {
	c.BaseSource = baseSource
	c.RevisionSource = revisionSource
	return c
}

func (c InfoChange) GetSection() string {
	return "info"
}

func (c InfoChange) IsBreaking() bool {
	return c.GetLevel().IsBreaking()
}

func (c InfoChange) MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool {
	return strings.Contains(ignoreLine, strings.ToLower(c.GetUncolorizedText(l))) &&
		strings.Contains(ignoreLine, "info")
}

func (c InfoChange) GetId() string {
	return c.Id
}

func (c InfoChange) GetText(l Localizer) string {
	return l(c.Id, colorizedValues(c.Args)...)
}

func (c InfoChange) GetArgs() []any {
	return c.Args
}

func (c InfoChange) GetUncolorizedText(l Localizer) string {
	return l(c.Id, quotedValues(c.Args)...)
}

func (c InfoChange) GetComment(l Localizer) string {
	return l(c.Comment)
}

func (InfoChange) GetDisclaimers() []Disclaimer {
	return nil
}

func (c InfoChange) GetLevel() Level {
	return c.Level
}

func (InfoChange) GetOperation() string {
	return ""
}

func (InfoChange) GetOperationId() string {
	return ""
}

func (InfoChange) GetPath() string {
	return ""
}

func (InfoChange) GetSource() string {
	return ""
}

func (InfoChange) GetSourceFile() string {
	return ""
}

func (InfoChange) GetSourceLine() int {
	return 0
}

func (InfoChange) GetSourceLineEnd() int {
	return 0
}

func (InfoChange) GetSourceColumn() int {
	return 0
}

func (InfoChange) GetSourceColumnEnd() int {
	return 0
}

func (c InfoChange) SingleLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s, %s info %s [%s]. %s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), l("in"), c.GetText(l), color.InYellow(c.Id), c.GetComment(l))
	}
	return fmt.Sprintf(format, c.Level.String(), l("in"), c.GetUncolorizedText(l), c.Id, c.GetComment(l))
}

func (c InfoChange) MultiLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s\t[%s]\n\t%s info\n\t\t%s%s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), color.InYellow(c.Id), l("in"), c.GetText(l), multiLineComment(c.GetComment(l)))
	}

	return fmt.Sprintf(format, c.Level.String(), c.Id, l("in"), c.GetUncolorizedText(l), multiLineComment(c.GetComment(l)))
}
