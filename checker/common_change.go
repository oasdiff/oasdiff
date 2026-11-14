package checker

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
