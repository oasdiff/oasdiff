package checker

import "github.com/oasdiff/oasdiff/diff"

// isNullableWrapping reports the nullable oneOf wrapping (see
// diff.NullableWrappingDiff). Used only by the transition's reporters (the
// became-nullable checks) to decide emission; all other checks are covered by
// the claim filter (see transition_claims.go).
func isNullableWrapping(schemaDiff *diff.SchemaDiff) bool {
	return schemaDiff != nil && !schemaDiff.NullableWrappingDiff.Empty()
}
