package checker

import (
	"strings"
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

func (c InfoChange) GetArgs() []any {
	return c.Args
}

// Text is the change's message, with each argument set apart by format.
func (c InfoChange) Text(l Localizer, format ArgFormat) string {
	return l(c.Id, formattedValues(c.Args, format)...)
}

func (c InfoChange) GetUncolorizedText(l Localizer) string {
	return c.Text(l, Quoted)
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

// Subject is the info object, which a change there is about as a whole.
func (c InfoChange) Subject() (section, element string) {
	return "info", ""
}

func (InfoChange) GetPath() string {
	return ""
}
