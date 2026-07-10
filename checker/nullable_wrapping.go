package checker

import "github.com/oasdiff/oasdiff/diff"

// isNullableWrapping reports that a schema was made nullable by the oneOf
// wrapping (see diff.NullableWrappingDiff). Used by the transition's
// reporters (the became-nullable checks) to decide emission; all other checks
// are covered by the claim filter (see transition_claims.go).
func isNullableWrapping(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil && schemaDiff.NullableWrappingDiff != nil && schemaDiff.NullableWrappingDiff.NullabilityAdded
}

// isNullableUnwrapping reports the reverse: the wrapper was removed, so the
// schema is no longer nullable. Used by the became-not-nullable reporters.
func isNullableUnwrapping(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil && schemaDiff.NullableWrappingDiff != nil && schemaDiff.NullableWrappingDiff.NullabilityRemoved
}
