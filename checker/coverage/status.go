package coverage

// Status is what the audit decided about one possible edit.
type Status string

const (
	// Covered: at least one check claims the edit.
	Covered Status = "covered"
	// Uncovered: a wire-relevant edit with no check and no waiver;
	// the audit fails the build until it gains one or the other.
	Uncovered Status = "uncovered"
	// Waived: a wire-relevant edit with no check, accounted for by a
	// coverage waiver.
	Waived Status = "waived"
	// NonContract: the edit cannot change which payloads are valid
	// (an annotation, a specification extension, or a metaschema.NonContracts
	// entry), so no check is expected.
	NonContract Status = "non-contract"
)
