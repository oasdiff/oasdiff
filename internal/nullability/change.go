package nullability

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/internal/oasversion"
)

// Transition is how a schema's acceptance of null changed.
type Transition int

const (
	Unchanged Transition = iota
	BecameNullable
	BecameNotNullable
)

// Change compares what Accepts says of the two versions of a schema, each
// read in its own document's dialect, and reports whether both could be
// decided.
func Change(base, revision *openapi3.Schema, dialects oasversion.Dialects) (transition Transition, decided bool) {
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
