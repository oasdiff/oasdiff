package formatters

import (
	"fmt"

	"github.com/TwiN/go-color"
	"github.com/oasdiff/oasdiff/checker"
)

// singleLineError renders a change as one line, and multiLineError over three.
// Both read the same parts from the change: what it is about, where it is, and
// what it says.
func singleLineError(change checker.Change, l checker.Localizer, colorMode checker.ColorMode) string {
	const format = "%s%s, %s %s %s [%s]. %s"

	return fmt.Sprintf(format,
		levelText(change.GetLevel(), colorMode),
		at(change, l),
		l("in"),
		subjectText(change, colorMode),
		changeText(change, l, colorMode),
		idText(change.GetId(), colorMode),
		change.GetComment(l))
}

func multiLineError(change checker.Change, l checker.Localizer, colorMode checker.ColorMode) string {
	const format = "%s\t[%s]%s\n\t%s %s\n\t\t%s%s"

	return fmt.Sprintf(format,
		levelText(change.GetLevel(), colorMode),
		idText(change.GetId(), colorMode),
		at(change, l),
		l("in"),
		subjectText(change, colorMode),
		changeText(change, l, colorMode),
		multiLineComment(change.GetComment(l)))
}

// levelText names a level, coloured by severity.
func levelText(level checker.Level, colorMode checker.ColorMode) string {
	name := level.String()
	if !checker.IsColorEnabled(colorMode) {
		return name
	}

	switch level {
	case checker.ERR:
		return color.InRed(name)
	case checker.WARN:
		return color.InPurple(name)
	case checker.INFO:
		return color.InCyan(name)
	}
	return color.InGray(name)
}

func idText(id string, colorMode checker.ColorMode) string {
	if checker.IsColorEnabled(colorMode) {
		return color.InYellow(id)
	}
	return id
}

// subjectText names what the change is about. The section reads plain and the
// element within it is coloured, as that is the part a reader scans for.
func subjectText(change checker.Change, colorMode checker.ColorMode) string {
	section, element := change.Subject()
	if element == "" {
		return section
	}

	if checker.IsColorEnabled(colorMode) {
		element = color.InGreen(element)
	}
	return section + " " + element
}

// changeText is the change's localized message, with its arguments quoted for
// a plain report and emphasised for a coloured one.
func changeText(change checker.Change, l checker.Localizer, colorMode checker.ColorMode) string {
	if !checker.IsColorEnabled(colorMode) {
		return change.GetUncolorizedText(l)
	}
	return change.Text(l, emphasized)
}

// emphasized sets an argument apart in a coloured report.
func emphasized(arg string) string {
	return color.InBold(fmt.Sprintf("'%s'", arg))
}

// at names the file the change is in. The revision is the side being
// reviewed, and a side that does not have the element has no location, so a
// change the revision does not have falls back to the base and one from a
// spec with no file, such as stdin, gets no segment at all.
func at(change checker.Change, l checker.Localizer) string {
	source := change.GetRevisionSource()
	if source == nil {
		source = change.GetBaseSource()
	}
	if source == nil || source.File == "" {
		return ""
	}
	return fmt.Sprintf(" %s %s", l("at"), source.File)
}

func multiLineComment(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("\n\t\t%s", comment)
}
