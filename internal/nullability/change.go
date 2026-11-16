package nullability

import "github.com/getkin/kin-openapi/openapi3"

// Transition is how a schema's acceptance of null changed.
type Transition int

const (
	Unchanged Transition = iota
	BecameNullable
	BecameNotNullable
)

// Dialects are the dialects of the two documents being compared. They differ
// when a document moved from OpenAPI 3.0 to 3.1 between the two versions.
type Dialects struct {
	Base, Revision Dialect
}

// Change compares what Accepts says of the two versions of a schema, each
// read in its own document's dialect, and reports whether both could be
// decided.
func Change(base, revision *openapi3.Schema, dialects Dialects) (transition Transition, decided bool) {
	before, beforeDecided := Accepts(base, dialects.Base)
	after, afterDecided := Accepts(revision, dialects.Revision)
	switch {
	case !beforeDecided || !afterDecided:
		return Unchanged, false
	case before == after:
		return Unchanged, true
	case after:
		return BecameNullable, true
	default:
		return BecameNotNullable, true
	}
}
