package checker

import "github.com/oasdiff/oasdiff/checker/internal/location"

// Source is where a change is in a spec. It is defined with the functions
// that find it, and named here because the change types hold it.
type Source = location.Source

type Change interface {
	GetSection() string
	IsBreaking() bool
	GetId() string
	GetArgs() []any

	// Text is the change's message, with each argument set apart by format.
	// GetUncolorizedText is the plain form of it.
	Text(l Localizer, format ArgFormat) string
	GetUncolorizedText(l Localizer) string
	GetComment(l Localizer) string
	GetLevel() Level

	GetDisclaimers() []Disclaimer
	GetOperation() string
	GetOperationId() string
	GetPath() string

	// Subject is what the change is about: the section it is in, and the
	// element within it, which a report may set apart. The element is empty
	// for a change about a section as a whole.
	Subject() (section, element string)
	GetAttributes() map[string]any

	// Location tracking methods
	GetBaseSource() *Source
	GetRevisionSource() *Source
	MatchIgnore(ignorePath, ignoreLine string, l Localizer) bool
}
