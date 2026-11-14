package checker

import (
	"strings"
)

// SecurityChange represents a change in the Security Section (not to be confised with components/securitySchemes)
type SecurityChange struct {
	CommonChange

	Id      string
	Args    []any
	Comment string
	Level   Level
}

// WithSources returns a copy of the SecurityChange with BaseSource and RevisionSource populated
func (c SecurityChange) WithSources(baseSource, revisionSource *Source) SecurityChange {
	c.BaseSource = baseSource
	c.RevisionSource = revisionSource
	return c
}

func (c SecurityChange) GetSection() string {
	return "security"
}

func (c SecurityChange) IsBreaking() bool {
	return c.GetLevel().IsBreaking()
}

func (c SecurityChange) MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool {
	return strings.Contains(ignoreLine, strings.ToLower(c.GetUncolorizedText(l))) &&
		strings.Contains(ignoreLine, "security")
}

func (c SecurityChange) GetId() string {
	return c.Id
}

func (c SecurityChange) GetArgs() []any {
	return c.Args
}

// Text is the change's message, with each argument set apart by format.
func (c SecurityChange) Text(l Localizer, format ArgFormat) string {
	return l(c.Id, formattedValues(c.Args, format)...)
}

func (c SecurityChange) GetUncolorizedText(l Localizer) string {
	return c.Text(l, Quoted)
}

func (c SecurityChange) GetComment(l Localizer) string {
	return l(c.Comment)
}

func (SecurityChange) GetDisclaimers() []Disclaimer {
	return nil
}

func (c SecurityChange) GetLevel() Level {
	return c.Level
}

func (r SecurityChange) GetOperation() string {
	return ""
}

func (SecurityChange) GetOperationId() string {
	return ""
}

// Subject is the security section, which a change there is about as a whole.
func (c SecurityChange) Subject() (section, element string) {
	return "security", ""
}

func (SecurityChange) GetPath() string {
	return ""
}
