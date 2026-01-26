package checker

import (
	"fmt"
	"strings"

	"github.com/TwiN/go-color"
)

// SecurityChange represents a change in the Security Section (not to be confised with components/securitySchemes)
type SecurityChange struct {
	CommonChange

	Id      string
	Args    []any
	Comment string
	Level   Level

	SourceFile      string
	SourceLine      int
	SourceLineEnd   int
	SourceColumn    int
	SourceColumnEnd int
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

func (c SecurityChange) GetText(l Localizer) string {
	return l(c.Id, colorizedValues(c.Args)...)
}

func (c SecurityChange) GetArgs() []any {
	return c.Args
}

func (c SecurityChange) GetUncolorizedText(l Localizer) string {
	return l(c.Id, quotedValues(c.Args)...)
}

func (c SecurityChange) GetComment(l Localizer) string {
	return l(c.Comment)
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

func (SecurityChange) GetPath() string {
	return ""
}

func (c SecurityChange) GetSource() string {
	return ""
}

func (c SecurityChange) GetSourceFile() string {
	return c.SourceFile
}

func (c SecurityChange) GetSourceLine() int {
	return c.SourceLine
}

func (c SecurityChange) GetSourceLineEnd() int {
	return c.SourceLineEnd
}

func (c SecurityChange) GetSourceColumn() int {
	return c.SourceColumn
}

func (c SecurityChange) GetSourceColumnEnd() int {
	return c.SourceColumnEnd
}

func (c SecurityChange) SingleLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s, %s security %s [%s]. %s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), l("in"), c.GetText(l), color.InYellow(c.Id), c.GetComment(l))
	}
	return fmt.Sprintf(format, c.Level.String(), l("in"), c.GetUncolorizedText(l), c.Id, c.GetComment(l))
}

func (c SecurityChange) MultiLineError(l Localizer, colorMode ColorMode) string {
	const format = "%s\t[%s] \t\n\t%s security\n\t\t%s%s"

	if isColorEnabled(colorMode) {
		return fmt.Sprintf(format, c.Level.PrettyString(), color.InYellow(c.Id), l("in"), c.GetText(l), multiLineComment(c.GetComment(l)))
	}

	return fmt.Sprintf(format, c.Level.String(), c.Id, l("in"), c.GetUncolorizedText(l), multiLineComment(c.GetComment(l)))
}
