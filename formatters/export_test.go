package formatters

import "github.com/oasdiff/oasdiff/checker"

// The renderers are the package's own business; these let its tests reach
// them without making them part of the API.
func SingleLineErrorForTest(change checker.Change, l checker.Localizer, colorMode checker.ColorMode) string {
	return singleLineError(change, l, colorMode)
}

func MultiLineErrorForTest(change checker.Change, l checker.Localizer, colorMode checker.ColorMode) string {
	return multiLineError(change, l, colorMode)
}

func LevelTextForTest(level checker.Level, colorMode checker.ColorMode) string {
	return levelText(level, colorMode)
}
