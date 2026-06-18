package checker

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

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
	if isRequestTypeWidening(typeDiff, schemaDiff) {
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
	if isResponseTypeNarrowing(typeDiff, schemaDiff) {
		typeDiff = nil
	}
	return typeFormatBreaking(typeDiff.Reverse(), formatDiff.Reverse(), isStronglyTyped(mediaType), schemaDiff.Base.Type)
}

// isRequestTypeWidening reports whether a request type change widens the accepted
// type set: the base set is contained in the revision set, so the server accepts
// every type it did before plus possibly more. Backward compatible for a request.
func isRequestTypeWidening(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return typeDiff != nil && isTypeSetSubset(getBaseType(schemaDiff), getRevisionType(schemaDiff))
}

// isResponseTypeNarrowing reports whether a response type change narrows the
// returned type set: the revision set is contained in the base set, so the server
// returns only types it could before. Backward compatible for a response. Mirror
// of isRequestTypeWidening.
func isResponseTypeNarrowing(typeDiff *diff.StringsDiff, schemaDiff *diff.SchemaDiff) bool {
	return typeDiff != nil && isTypeSetSubset(getRevisionType(schemaDiff), getBaseType(schemaDiff))
}

// isTypeSetSubset reports whether both sets are non-empty and every type in sub is
// covered by some type in super (using the integer-within-number lattice). An
// empty set means no type constraint (any value); that universe case is handled
// by the empty-target-type short-circuit in typeFormatBreaking, not here.
func isTypeSetSubset(sub, super []string) bool {
	if len(sub) == 0 || len(super) == 0 {
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
checkRequestParameterPropertyTypeChanged checks the level of the change in the request parameter property type
Explanation:
Objects can be passed in the request parameters, for example, the following calls are equivalent:
PHP style: GET http://localhost:8080/api/tickets?params[id]=123&params[color]=green
JSON: GET http://localhost:8080/api/tickets?params={"id":"123","color":"green"}

The "params" object has two properties: "id" and "color", both with type "string", but note that the "id" values are actually numbers.
Imagine that the OpenAPI type of property "id" was changed from "number" to "string".
In the first example, the change is non-breaking, because the PHP format for numbers and strings is the same: we refer to this as non-strongly-typed.
But in the second example, the change is breaking, because the JSON format requires quotes for strings: we refer to this as strongly-typed.

This is the only request type location that forks three ways
(generalized / specialized / changed-as-a-warning). The other request type
locations resolve strong-vs-non-strong definitively (the body and body
properties from a known media type; a scalar parameter is always a string on
the wire), so a binary generalized/changed verdict is correct there. Only an
object parameter's serialization is unknown here, so when the two verdicts
disagree we can't be sure it's breaking and report a warning.
*/
func checkRequestParameterPropertyTypeChanged(typeDiff *diff.StringsDiff, formatDiff *diff.ValueDiff, schemaDiff *diff.SchemaDiff) (string, string) {

	// since we don't know if the object is strogly-typed or not, we check both
	stronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, true, schemaDiff.Revision.Type)
	nonStronglyTyped := typeOrFormatBreaking(typeDiff, formatDiff, false, schemaDiff.Revision.Type)

	// if strongly-typed and non-strongly-typed don't agree, it's a warning since we can't be sure that it's breaking
	if stronglyTyped != nonStronglyTyped {
		return RequestParameterPropertyTypeChangedId, RequestParameterPropertyTypeChangedCommentId
	}

	// if both are breaking it's an error
	if stronglyTyped {
		return RequestParameterPropertyTypeSpecializedId, ""
	}

	// if neither are breaking it's an informational change
	return RequestParameterPropertyTypeGeneralizedId, ""
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

func getBaseFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Base.Format
}

func getRevisionFormat(schemaDiff *diff.SchemaDiff) string {
	return schemaDiff.Revision.Format
}
