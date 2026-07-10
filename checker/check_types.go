package checker

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// TypeChangeLooselyTypedCommentId explains the surprising "compatible" verdict:
// a type change that is safe only because the media type isn't strongly typed.
const TypeChangeLooselyTypedCommentId = "type-change-loosely-typed-comment"

// requestTypeFormatBreaking reports whether a request type/format change is
// breaking. Requests are contravariant, so the change is evaluated toward the
// revision type.
//
// A type-set widening (the revision accepts every type the base did, plus
// possibly more) is backward compatible on the type axis, but the diff-based
// core only sees the added/removed types and would still flag a multi-type
// widening. We detect it from the full type sets and drop the type axis; the
// format axis is evaluated independently, so a co-occurring breaking format
// change is still reported.
func requestTypeFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {
	if typeSetWidened(typeDiff, schemaDiff) {
		typeDiff = nil
	}
	return typeFormatBreaking(typeDiff, formatDiff, isStronglyTyped(mediaType), schemaDiff.Revision.Type)
}

// responseTypeFormatBreaking reports whether a response type/format change is
// breaking. Responses are covariant, so it is the same check with the
// base/revision direction reversed: the diffs are reversed and the change is
// evaluated toward the base type.
//
// The mirror of the request widening case: a type-set narrowing (the revision
// returns only types the base could) is backward compatible on the type axis,
// and is detected and dropped the same way.
func responseTypeFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, mediaType string, schemaDiff *diff.SchemaDiff) bool {
	if typeSetNarrowed(typeDiff, schemaDiff) {
		typeDiff = nil
	}
	return typeFormatBreaking(typeDiff.Reverse(), formatDiff.Reverse(), isStronglyTyped(mediaType), schemaDiff.Base.Type)
}

// typeSetWidened reports whether the change added types (base is contained in
// revision). Safe for a request (accepts more), breaking for a response (may
// return more); the caller decides which.
func typeSetWidened(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return typeDiff != nil && isTypeSetSubset(getBaseType(schemaDiff), getRevisionType(schemaDiff))
}

// typeSetNarrowed is the mirror: the change removed types (revision is contained
// in base). Safe for a response, breaking for a request.
func typeSetNarrowed(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return typeDiff != nil && isTypeSetSubset(getRevisionType(schemaDiff), getBaseType(schemaDiff))
}

// bothTypesConcrete distinguishes a type swap (string -> object) from adding or
// removing the type constraint entirely, which is a real generalize/specialize.
func bothTypesConcrete(schemaDiff *diff.SchemaDiff) bool {
	return len(getBaseType(schemaDiff)) > 0 && len(getRevisionType(schemaDiff)) > 0
}

// isRequestLooseTypeSwap reports a concrete type swap that is not a genuine
// widening. Its caller reaches it only when the change is non-breaking, so the
// swap is safe only because the media type isn't strongly typed.
func isRequestLooseTypeSwap(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return !typeDiff.Empty() && bothTypesConcrete(schemaDiff) && !typeSetWidened(typeDiff, schemaDiff)
}

// isResponseLooseTypeSwap is the response mirror: a concrete swap that is not a
// genuine narrowing.
func isResponseLooseTypeSwap(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return !typeDiff.Empty() && bothTypesConcrete(schemaDiff) && !typeSetNarrowed(typeDiff, schemaDiff)
}

// isTypeSetSubset reports whether sub is non-empty and every type in it is
// covered by some type in super (using the integer-within-number lattice).
//
// An empty sub is treated as not-a-subset rather than the vacuous "empty set is
// a subset of anything": that keeps the empty-source cases (a type removed from
// a response, or a constraint added to a previously untyped request) on the
// breaking path. The empty-target case (no type constraint = any value) is
// handled by the empty-target-type short-circuit in typeFormatBreaking, not here.
func isTypeSetSubset(sub, super []string) bool {
	if len(sub) == 0 {
		return false
	}
	for _, t := range sub {
		if !typeCoveredBy(t, super) {
			return false
		}
	}
	return true
}

// isSubtype reports whether every value of type sub is also a valid value of
// type super: the same type, or "integer" within "number" (every integer is a
// number). This is the single source of the type-compatibility lattice, shared
// by isTypeContained (diff-shaped) and typeCoveredBy (set-membership).
func isSubtype(sub, super string) bool {
	return sub == super || (sub == "integer" && super == "number")
}

// typeCoveredBy reports whether type t is covered by one of the base types (t is
// a subtype of some base type).
func typeCoveredBy(t string, base []string) bool {
	return slices.ContainsFunc(base, func(b string) bool { return isSubtype(t, b) })
}

// typeFormatBreaking reports whether a type/format change toward toType is
// breaking. An empty toType means the type constraint is gone: removing it on
// the request side (the server then accepts any value), or, in the reversed
// response frame, adding one (the server then returns a subset). Both are
// non-breaking generalizations. Otherwise the two axes are evaluated by the
// direction-agnostic core.
func typeFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, stronglyTyped bool, toType *openapi3.Types) bool {
	if typeDiff != nil && len(toType.Slice()) == 0 {
		return false
	}
	return typeOrFormatBreaking(typeDiff, formatDiff, stronglyTyped, toType)
}

// typeOrFormatBreaking reports whether a type change or a format change is
// breaking, evaluating the two axes independently toward toType: the change is
// breaking if either the type or the format is breaking on its own. It does not
// treat a removed type constraint specially; callers that need that go through
// typeFormatBreaking.
// stronglyTyped reflects the media type (see isStronglyTyped); callers that
// can't resolve it (request parameters) pass it explicitly.
func typeOrFormatBreaking(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, stronglyTyped bool, toType *openapi3.Types) bool {
	typeBreaking := typeDiff != nil && !isTypeContained(typeDiff.Added, typeDiff.Deleted, stronglyTyped)
	formatBreaking := formatDiff != nil && !isFormatContained(toType, formatDiff.To, formatDiff.From)
	return typeBreaking || formatBreaking
}

/*
isTypeContained checks if type2 is contained in type1
note that we don't support multiple types currenty
*/
func isTypeContained(to, from []string, stronglyTyped bool) bool {

	// from (the old type) is a subtype of to (the new type): the new type still
	// accepts every value the old one did. Added and Deleted are disjoint, so
	// to[0] == from[0] cannot occur here, leaving the integer-within-number case.
	if len(to) == 1 && len(from) == 1 && isSubtype(from[0], to[0]) {
		return true
	}

	// anything can be changed to string, unless it's "strongly typed"
	if !stronglyTyped {
		return len(to) == 0 || (len(to) == 1 && to[0] == "string")
	}

	return false
}

/*
isStronglyTyped checks if the media type is strongly typed, for example:
in text format, all numbers can also be interpreted as strings (1 can be a number or a string)
but in json, a number (1) is not the same as a string ("1")
*/
func isStronglyTyped(mediaType string) bool {
	return isJsonMediaType(mediaType)
}

func isJsonMediaType(mediaType string) bool {
	// Structured Syntax Suffixes: https://www.rfc-editor.org/rfc/rfc6838#section-4.2.8
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

// isFormatContained checks if from is contained in to
func isFormatContained(revisionType *openapi3.Types, to, from any) bool {

	// Removing a format constraint is a generalization (non-breaking), whatever
	// the type, including when the type was removed too (revisionType is nil).
	// Checked before the type switch so it applies for any revision type.
	if to == "" {
		return true
	}

	if revisionType == nil || len(*revisionType) > 1 {
		return false
	}

	// we don't support multiple types currenty, so just take the first one
	switch getSingleType(revisionType) {
	case "number":
		return to == "double" && from == "float"
	case "integer":
		return (to == "int64" && from == "int32") ||
			(to == "bigint" && from == "int32") ||
			(to == "bigint" && from == "int64")
	case "string":
		return (to == "date-time" && from == "date") ||
			(to == "date-time" && from == "time")
	}

	return false
}

func getSingleType(types *openapi3.Types) string {
	if types == nil || len(*types) == 0 {
		return ""
	}

	return (*types)[0]
}

func getBaseType(schemaDiff *diff.SchemaDiff) []string {
	return schemaDiff.Base.Type.Slice()
}

func getRevisionType(schemaDiff *diff.SchemaDiff) []string {
	return schemaDiff.Revision.Type.Slice()
}

// type/format rendering for human-readable messages lives in type_format.go.
