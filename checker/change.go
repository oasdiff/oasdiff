package checker

import (
	"fmt"

	"github.com/oasdiff/oasdiff/checker/internal/location"
)

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

// source is where the change is reported. A side that does not have the
// element has no location, so the order decides only a change that is in both,
// and there the revision is the side being reviewed.
func (c CommonChange) source() *Source {
	if c.RevisionSource != nil {
		return c.RevisionSource
	}
	return c.BaseSource
}

// at renders the "at <file>" segment of an error line, empty when the change
// has no file to name, as a spec read from stdin has none.
func (c CommonChange) at(l Localizer) string {
	source := c.source()
	if source == nil || source.File == "" {
		return ""
	}
	return fmt.Sprintf(" %s %s", l("at"), source.File)
}
