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

// source is where the change is reported. A side that does not have the
// element has no location, so the order decides only a change that is in both,
// and there the revision is the side being reviewed.
func (c InfoChange) source() *Source {
	if c.RevisionSource != nil {
		return c.RevisionSource
	}
	return c.BaseSource
}

func (c InfoChange) GetSource() string {
	if source := c.source(); source != nil {
		return source.File
	}
	return ""
}

// at renders the "at <file>" segment of an error line, empty when the change
// has no location.
func (c InfoChange) at(l Localizer) string {
	file := c.GetSource()
	if file == "" {
		return ""
	}
	return fmt.Sprintf(" %s %s", l("at"), file)
}

func (c InfoChange) SingleLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s%s, %s info %s [%s]. %s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), c.at(l), l("in"), c.GetText(l), color.InYellow(c.Id), c.GetComment(l))
	}
	return fmt.Sprintf(format, c.Level.String(), c.at(l), l("in"), c.GetUncolorizedText(l), c.Id, c.GetComment(l))
}

func (c InfoChange) MultiLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s\t[%s]%s\n\t%s info\n\t\t%s%s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), color.InYellow(c.Id), c.at(l), l("in"), c.GetText(l), multiLineComment(c.GetComment(l)))
	}

	return fmt.Sprintf(format, c.Level.String(), c.Id, c.at(l), l("in"), c.GetUncolorizedText(l), multiLineComment(c.GetComment(l)))
}
