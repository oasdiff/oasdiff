package checker

import "github.com/oasdiff/oasdiff/checker/internal/location"

// Source is where a change is in a spec. It is defined with the functions
// that find it, and named here because the change types hold it.
type Source = location.Source

type Change interface {
	GetSection() string
	IsBreaking() bool
	GetId() string
	GetText(l Localizer) string
	GetArgs() []any
	GetUncolorizedText(l Localizer) string
	GetComment(l Localizer) string
	GetLevel() Level

	GetDisclaimers() []Disclaimer
	GetOperation() string
	GetOperationId() string
	GetPath() string
	GetAttributes() map[string]any

	// Location tracking methods
	GetBaseSource() *Source
	GetRevisionSource() *Source
	MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool
	SingleLineError(l Localizer, colorMode ColorMode) string
	MultiLineError(l Localizer, colorMode ColorMode) string
}
