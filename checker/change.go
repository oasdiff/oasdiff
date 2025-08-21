package checker

type Change interface {
	GetSection() string
	IsBreaking() bool
	GetId() string
	GetText(l Localizer) string
	GetArgs() []any
	GetUncolorizedText(l Localizer) string
	GetComment(l Localizer) string
	GetLevel() Level
	GetOperation() string
	GetOperationId() string
	GetPath() string
	GetSource() string
	GetAttributes() map[string]any

	// Location tracking methods
	GetBaseSource() *Source
	GetRevisionSource() *Source

	// DEPRECATED: Will be removed after GitHubActionsFormatter migration
	GetSourceFile() string
	GetSourceLine() int
	GetSourceLineEnd() int
	GetSourceColumn() int
	GetSourceColumnEnd() int
	MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool
	SingleLineError(l Localizer, colorMode ColorMode) string
	MultiLineError(l Localizer, colorMode ColorMode) string
}

type CommonChange struct {
	// Location information for correlation UI
	BaseSource     *Source // Location in base (original) file
	RevisionSource *Source // Location in revision (modified) file

	Attributes map[string]any
}

func (c CommonChange) GetBaseSource() *Source {
	return c.BaseSource
}

func (c CommonChange) GetRevisionSource() *Source {
	return c.RevisionSource
}

func (c CommonChange) GetAttributes() map[string]any {
	return c.Attributes
}
