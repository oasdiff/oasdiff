package checker

import "fmt"

// CommonChange is what every change type holds: where the change is on each
// side, and the extensions the caller asked to carry along.
type CommonChange struct {
	BaseSource     *Source
	RevisionSource *Source

	// Attributes holds the extensions named by Config.Attributes that the
	// operation this change was found on actually has.
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
