package checker

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// isNullTypeChange returns true if the only change in a type list is adding or removing "null"
func isNullTypeChange(typeDiff *diff.StringsDiff) bool {
	if typeDiff == nil {
		return false
	}
	return onlyNull(typeDiff.Added) && onlyNull(typeDiff.Deleted)
}

// nullAddedToTypeArray returns true if "null" was added to the type array and the
// base already constrained the type (OpenAPI 3.1 became nullable). If the base had
// no type keyword, it was untyped and already accepted any value, including null,
// so introducing an explicit type that contains null does NOT make it newly
// nullable, even though "null" appears in the added set. baseType is the base
// schema's type. Mirror of nullRemovedFromTypeArray; see #1004.
func nullAddedToTypeArray(typeDiff *diff.StringsDiff, baseType *openapi3.Types) bool {
	if typeDiff == nil {
		return false
	}
	if baseType == nil || len(*baseType) == 0 {
		return false
	}
	return slices.Contains(typeDiff.Added, "null")
}

// nullRemovedFromTypeArray returns true if "null" was removed from the type array
// and the revision still constrains the type (OpenAPI 3.1 became not-nullable).
// If the revision dropped the type keyword entirely, the schema is untyped and
// accepts any value, including null, so it did NOT become non-nullable, even
// though "null" appears in the deleted set. revisionType is the revision
// schema's type. See #1004.
func nullRemovedFromTypeArray(typeDiff *diff.StringsDiff, revisionType *openapi3.Types) bool {
	if typeDiff == nil {
		return false
	}
	if revisionType == nil || len(*revisionType) == 0 {
		return false
	}
	return slices.Contains(typeDiff.Deleted, "null")
}

func onlyNull(types []string) bool {
	for _, t := range types {
		if t != "null" {
			return false
		}
	}
	return true
}

// nullability classifies a schema diff's nullability transition.
type nullability int

const (
	nullabilityUnchanged nullability = iota
	becameNullable
	becameNotNullable
)

// nullabilityChangeId returns the rule id reporting the schema diff's
// nullability transition, or "" when nullability did not change.
func nullabilityChangeId(d *diff.SchemaDiff, nullableId, notNullableId string) string {
	switch nullabilityChange(d) {
	case becameNullable:
		return nullableId
	case becameNotNullable:
		return notNullableId
	}
	return ""
}

// nullabilityChange recognizes a nullability transition in any of its three
// equivalent forms: the nullable keyword (OpenAPI 3.0), a "null" entry in the
// type array (OpenAPI 3.1), and the nullable oneOf wrapping
// (oneOf: [{type: "null"}, <equivalent schema>]).
func nullabilityChange(d *diff.SchemaDiff) nullability {
	if d.NullableDiff != nil {
		if d.NullableDiff.From == true {
			return becameNotNullable
		}
		if d.NullableDiff.To == true {
			return becameNullable
		}
		return nullabilityUnchanged
	}
	if nullRemovedFromTypeArray(d.TypeDiff, d.Revision.Type) {
		return becameNotNullable
	}
	if nullAddedToTypeArray(d.TypeDiff, d.Base.Type) {
		return becameNullable
	}
	if isNullableWrapping(d) {
		return becameNullable
	}
	if isNullableUnwrapping(d) {
		return becameNotNullable
	}
	return nullabilityUnchanged
}
