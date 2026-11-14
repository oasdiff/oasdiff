package checker

import (
	"strings"
)

// ComponentChange represnts a change in the Components Section: https://swagger.io/docs/specification/components/
type ComponentChange struct {
	CommonChange

	Id        string
	Args      []any
	Comment   string
	Level     Level
	Component string
}

// WithSources returns a copy of the ComponentChange with BaseSource and RevisionSource populated
func (c ComponentChange) WithSources(baseSource, revisionSource *Source) ComponentChange {
	c.BaseSource = baseSource
	c.RevisionSource = revisionSource
	return c
}

func (c ComponentChange) GetSection() string {
	return "components"
}

func (c ComponentChange) IsBreaking() bool {
	return c.GetLevel().IsBreaking()
}

func (c ComponentChange) MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool {
	return strings.Contains(ignoreLine, strings.ToLower(c.GetUncolorizedText(l))) &&
		strings.Contains(ignoreLine, "components")
}

func (c ComponentChange) GetId() string {
	return c.Id
}

func (c ComponentChange) GetArgs() []any {
	return c.Args
}

// Text is the change's message, with each argument set apart by format.
func (c ComponentChange) Text(l Localizer, format ArgFormat) string {
	return l(c.Id, formattedValues(c.Args, format)...)
}

func (c ComponentChange) GetUncolorizedText(l Localizer) string {
	return c.Text(l, Quoted)
}

func (c ComponentChange) GetComment(l Localizer) string {
	return l(c.Comment)
}

func (ComponentChange) GetDisclaimers() []Disclaimer {
	return nil
}

func (c ComponentChange) GetLevel() Level {
	return c.Level
}

func (ComponentChange) GetOperation() string {
	return ""
}

func (ComponentChange) GetOperationId() string {
	return ""
}

// Subject is the kind of component the change was found in.
func (c ComponentChange) Subject() (section, element string) {
	return "components/" + c.Component, ""
}

func (ComponentChange) GetPath() string {
	return ""
}
