package nullability

import "github.com/getkin/kin-openapi/openapi3"

// Transition is how a schema's acceptance of null changed.
type Transition int

const (
	Unchanged Transition = iota
	BecameNullable
	BecameNotNullable
)

// Change compares what Accepts says of the two versions of a schema, and
// reports whether both could be decided.
func Change(base, revision *openapi3.Schema) (transition Transition, decided bool) {
	before, beforeDecided := Accepts(base)
	after, afterDecided := Accepts(revision)
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
