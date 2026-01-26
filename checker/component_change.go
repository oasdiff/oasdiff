package checker

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
)

// ComponentChange represnts a change in the Components Section: https://swagger.io/docs/specification/components/
type ComponentChange struct {
	CommonChange

	Id        string
	Args      []any
	Comment   string
	Level     Level
	Component string

	SourceFile      string
	SourceLine      int
	SourceLineEnd   int
	SourceColumn    int
	SourceColumnEnd int
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

func (c ComponentChange) GetText(l Localizer) string {
	return l(c.Id, colorizedValues(c.Args)...)
}

func (c ComponentChange) GetArgs() []any {
	return c.Args
}

func (c ComponentChange) GetUncolorizedText(l Localizer) string {
	return l(c.Id, quotedValues(c.Args)...)
}

func (c ComponentChange) GetComment(l Localizer) string {
	return l(c.Comment)
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

func (ComponentChange) GetPath() string {
	return ""
}

func (c ComponentChange) GetSource() string {
	return ""
}

func (c ComponentChange) GetSourceFile() string {
	return c.SourceFile
}

func (c ComponentChange) GetSourceLine() int {
	return c.SourceLine
}

func (c ComponentChange) GetSourceLineEnd() int {
	return c.SourceLineEnd
}

func (c ComponentChange) GetSourceColumn() int {
	return c.SourceColumn
}

func (c ComponentChange) GetSourceColumnEnd() int {
	return c.SourceColumnEnd
}

func (c ComponentChange) SingleLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s, %s components/%s %s [%s]. %s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), l("in"), c.Component, c.GetText(l), color.InYellow(c.Id), c.GetComment(l))
	}
	return fmt.Sprintf(format, c.Level.String(), l("in"), c.Component, c.GetUncolorizedText(l), c.Id, c.GetComment(l))
}

func (c ComponentChange) MultiLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s\t[%s] \t\n\t%s components/%s\n\t\t%s%s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), color.InYellow(c.Id), l("in"), c.Component, c.GetText(l), multiLineComment(c.GetComment(l)))
	}
	return fmt.Sprintf(format, c.Level.String(), c.Id, l("in"), c.Component, c.GetUncolorizedText(l), multiLineComment(c.GetComment(l)))
}
