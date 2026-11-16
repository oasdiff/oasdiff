package checker

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/internal/nullability"
)

// isNullTypeChange returns true if the only change in a type list is adding or removing "null"
func isNullTypeChange(typeDiff *diff.StringsDiff) bool {
	return typeDiff != nil && nullability.OnlyNullChanged(typeDiff.Added, typeDiff.Deleted)
}

// nullAddedToTypeArray reports whether "null" was added to a type array that
// already constrained the type (OpenAPI 3.1 became nullable). baseType is the
// base schema's type. See #1004.
func nullAddedToTypeArray(typeDiff *diff.StringsDiff, baseType *openapi3.Types) bool {
	return typeDiff != nil && nullability.AddedToTypes(typeDiff.Added, baseType.Slice())
}

// nullRemovedFromTypeArray reports whether "null" was removed from a type
// array that still constrains the type (OpenAPI 3.1 became not-nullable).
// revisionType is the revision schema's type. See #1004.
func nullRemovedFromTypeArray(typeDiff *diff.StringsDiff, revisionType *openapi3.Types) bool {
	return typeDiff != nil && nullability.RemovedFromTypes(typeDiff.Deleted, revisionType.Slice())
}

// nullabilityChangeId returns the rule id and comment reporting the schema
// diff's nullability transition, or "" when nullability did not change.
// parent is the schema diff d sits under, or nil at the root of a walk.
func nullabilityChangeId(parent *diff.SchemaDiff, d *diff.SchemaDiff, nullableId, notNullableId string) (string, string) {
	if branchOfComposition(parent, d) {
		return "", ""
	}
	transition, comment := nullabilityChange(d)
	switch transition {
	case nullability.BecameNullable:
		return nullableId, comment
	case nullability.BecameNotNullable:
		return notNullableId, comment
	}
	return "", comment
}

// nullabilityChange reports a schema's nullability transition and the comment
// to report it with. The nullable keyword (OpenAPI 3.0) and a "null" entry
// added to or removed from the type array (OpenAPI 3.1) are read as written.
// A schema with a oneOf or anyOf on either side, which includes one wrapped in
// oneOf: [{type: "null"}, <schema>], is decided by whether it accepted null
// before and after (composedNullability).
func nullabilityChange(d *diff.SchemaDiff) (nullability.Transition, string) {
	// A schema present on one side only is not a nullability transition (and
	// the type-array checks below read both sides).
	if d == nil || d.Base == nil || d.Revision == nil {
		return nullability.Unchanged, ""
	}
	if d.NullableDiff != nil {
		if d.NullableDiff.From == true {
			return nullability.BecameNotNullable, ""
		}
		if d.NullableDiff.To == true {
			return nullability.BecameNullable, ""
		}
		return nullability.Unchanged, ""
	}
	if nullRemovedFromTypeArray(d.TypeDiff, d.Revision.Type) {
		return nullability.BecameNotNullable, ""
	}
	if nullAddedToTypeArray(d.TypeDiff, d.Base.Type) {
		return nullability.BecameNullable, ""
	}
	if transition, comment, decided := composedNullability(d); decided {
		return transition, comment
	}
	// What one of the sides accepts is not decided, as when the wrapped schema
	// has an if: the wrapping itself is still recognized.
	if isNullableWrapping(d) {
		return nullability.BecameNullable, ""
	}
	if isNullableUnwrapping(d) {
		return nullability.BecameNotNullable, ""
	}
	return nullability.Unchanged, ""
}
